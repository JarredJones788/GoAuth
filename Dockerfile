FROM alpine:3.8

WORKDIR /usr/src/app

COPY ./logs ./logs
COPY ./server ./server

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

EXPOSE 4000
CMD ["/usr/src/app/server", "-production"]