FROM golang:alpine

RUN apk add socat git 

COPY ../ /root/slcan-svc

WORKDIR /root/slcan-svc

RUN go mod tidy \
    && go build .
