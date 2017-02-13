FROM nginx:latest

RUN rm -rf /etc/nginx/conf.d/*

COPY server.conf /etc/nginx/conf.d/server.conf
COPY nginx.conf /etc/nginx/nginx.conf

ENV TZ=Asia/Shanghai

EXPOSE 6003
EXPOSE 6004