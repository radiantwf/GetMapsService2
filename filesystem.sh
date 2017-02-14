#!/bin/bash
docker kill seaweedfs-cassandra seaweedfs-redis seaweedfs-master1 seaweedfs-volume1 seaweedfs-filer1 getmapservice;
docker rm seaweedfs-cassandra seaweedfs-redis seaweedfs-master1 seaweedfs-volume1 seaweedfs-filer1 getmapservice;
docker rmi 211.157.146.6:5000/getmapservice

docker run -d --name seaweedfs-redis 211.157.146.6:5000/redis
docker run -d -p 9333:9333 --name="seaweedfs-master1" 211.157.146.6:5000/seaweedfs master
docker run -d -p 8888:8080 --link="seaweedfs-master1:master" --name="seaweedfs-volume1" 211.157.146.6:5000/seaweedfs volume -max=20 -mserver="master:9333" -publicUrl="127.0.0.1:8888" -port=8080
docker run -d -p 8889:8080 --link="seaweedfs-master1:master" --link="seaweedfs-redis:redis" --name="seaweedfs-filer1" 211.157.146.6:5000/seaweedfs filer -redis.server="redis:6379" -master="master:9333" -port=8080
docker run -d --link "seaweedfs-filer1:filer" -p 8000:8000 --name="getmapservice" 211.157.146.6:5000/getmapservice

docker kill seaweedfs-cassandra seaweedfs-redis seaweedfs-master1 seaweedfs-volume1 seaweedfs-filer1 getmapservice;
docker rm seaweedfs-cassandra seaweedfs-redis seaweedfs-master1 seaweedfs-volume1 seaweedfs-filer1 getmapservice;
docker rmi 211.157.146.6:5000/getmapservice

docker run -d --name seaweedfs-cassandra 211.157.146.6:5000/cassandra
docker exec -it seaweedfs-cassandra 'cqlsh'
create keyspace seaweed WITH replication = {
  'class':'SimpleStrategy',
  'replication_factor':1
};

use seaweed;

CREATE TABLE seaweed_files (
   path varchar,
   fids list<varchar>,
   PRIMARY KEY (path)
);

docker run -d -p 9333:9333 --name="seaweedfs-master1" 211.157.146.6:5000/seaweedfs master
docker run -d -p 8888:8080 --link="seaweedfs-master1:master" --name="seaweedfs-volume1" 211.157.146.6:5000/seaweedfs volume -max=20 -mserver="master:9333" -publicUrl="127.0.0.1:8888" -port=8080
docker run -d -p 8889:8080 --link="seaweedfs-master1:master" --link="seaweedfs-cassandra:cassandra" --name="seaweedfs-filer1" 211.157.146.6:5000/seaweedfs filer -cassandra.server="cassandra" -master="master:9333" -port=8080
docker run -d --link "seaweedfs-filer1:filer" -p 8000:8000 --name="getmapservice" 211.157.146.6:5000/getmapservice

docker kill getmapservice;docker rm getmapservice;
docker rmi 211.157.146.6:5000/getmapservice
docker run -d --link "seaweedfs-filer1:filer" -p 8000:8000 --name="getmapservice" 211.157.146.6:5000/getmapservice

docker kill mapserver-nginx;docker rm mapserver-nginx;
docker rmi 211.157.146.6:5000/mapserver-nginx
docker run -d -p 6003:6003 -p 6004:6004 --name="mapserver-nginx" 211.157.146.6:5000/mapserver-nginx
