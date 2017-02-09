docker run -d --name seaweedfs-redis redis
docker run -d -p 9333:9333 --name="seaweedfs-master" chrislusf/seaweedfs master
docker run -d -p 8888:8080 --link="seaweedfs-master:master" --name="seaweedfs-volume1" chrislusf/seaweedfs volume -max=20 -mserver="master:9333" -publicUrl="127.0.0.1:8888" -port=8080
docker run -d -p 8889:8080 --link="seaweedfs-master:master" --link="seaweedfs-redis:redis" --name="seaweedfs-filer" chrislusf/seaweedfs filer -redis.server="redis:6379" -master="master:9333" -port=8080