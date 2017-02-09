FROM debian:latest

COPY ./resources/ /GetMapService/resources/
COPY ./main /GetMapService/main
ENV TZ=Asia/Shanghai
WORKDIR /GetMapService
CMD ["./main"]

EXPOSE 8000