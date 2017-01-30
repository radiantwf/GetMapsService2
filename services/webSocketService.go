package services

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketService 定义
type WebSocketService struct {
	addr              *string
	staticFilesHander http.Handler
	homeTempl         *template.Template
	submitCallback    SubmitCallback
	h                 hub
	Config            *ConfigService `inject:""`
}

// SubmitCallback 定义
type SubmitCallback func(message []byte)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 5120
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) readPump(h hub) {
	defer func() {
		h.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			continue
			// break
		}
		h.message <- message
	}
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func (service *WebSocketService) serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws}
	service.h.register <- c
	go c.writePump()
	c.readPump(service.h)
}

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	message chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

// run 定义
func (h *hub) run(service *WebSocketService) {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}
		case m := <-h.message:
			if service.submitCallback != nil {
				service.submitCallback(m)
			}
		}
	}
}

// serveHome 定义
func (service *WebSocketService) serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	service.homeTempl.Execute(w, r.Host)
}

// BroadcastMessage 定义
func (service *WebSocketService) BroadcastMessage(message string) {
	for c := range service.h.connections {
		select {
		case c.send <- []byte(message):
		default:
			close(c.send)
			delete(service.h.connections, c)
		}
	}
}

// Start 定义
func (service *WebSocketService) Start() {
	service.h = hub{
		message:     make(chan []byte),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
		connections: make(map[*connection]bool),
	}

	service.homeTempl = template.Must(template.ParseFiles(service.Config.configStruct.HTMLPath))
	service.staticFilesHander = http.FileServer(http.Dir(service.Config.configStruct.StaticPath))
	portStr := fmt.Sprintf(":%d", service.Config.configStruct.Port)
	service.addr = flag.String("addr", portStr, "http service address")

	go service.h.run(service)
	http.Handle("/static/", http.StripPrefix("/static/", service.staticFilesHander))
	http.HandleFunc("/", service.serveHome)
	http.HandleFunc("/ws", service.serveWs)
	err := http.ListenAndServe(*service.addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
