#!/bin/bash
docker rmi 211.157.146.6:5000/mapserver-nginx
docker build -t 211.157.146.6:5000/mapserver-nginx .
docker push 211.157.146.6:5000/mapserver-nginx
