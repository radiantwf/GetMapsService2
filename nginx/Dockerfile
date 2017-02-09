FROM nginx:latest

RUN set -x \
	&& apt-get update \
    && apt-get install -y --no-install-recommends \
    vim

RUN rm -rf /etc/nginx/conf.d/*

COPY server.conf /etc/nginx/conf.d/server.conf
COPY nginx.conf /etc/nginx/nginx.conf

ENV TZ=Asia/Shanghai

EXPOSE 6007
EXPOSE 6008