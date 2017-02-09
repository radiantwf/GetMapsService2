#!/bin/bash
docker rmi 211.157.146.6:5000/getmapservice
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build *.go
docker build -t 211.157.146.6:5000/getmapservice .
docker push 211.157.146.6:5000/getmapservice
