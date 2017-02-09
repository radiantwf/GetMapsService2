#!/bin/bash
docker kill seaweedfs-redis seaweedfs-master1 seaweedfs-volume1 seaweedfs-filer1 getmapservice;docker rm seaweedfs-redis seaweedfs-master1 seaweedfs-volume1 seaweedfs-filer1 getmapservice;

docker run -d --name seaweedfs-redis 211.157.146.6:5000/redis
docker run -d -p 9333:9333 --name="seaweedfs-master1" 211.157.146.6:5000/seaweedfs master
docker run -d -p 8888:8080 --link="seaweedfs-master1:master" --name="seaweedfs-volume1" 211.157.146.6:5000/seaweedfs volume -max=20 -mserver="master:9333" -publicUrl="127.0.0.1:8888" -port=8080
docker run -d -p 8889:8080 --link="seaweedfs-master1:master" --link="seaweedfs-redis:redis" --name="seaweedfs-filer1" 211.157.146.6:5000/seaweedfs filer -redis.server="redis:6379" -master="master:9333" -port=8080
docker run -d --link "seaweedfs-filer1:filer" -p 8000:8000 --name="getmapservice" 211.157.146.6:5000/getmapservice

