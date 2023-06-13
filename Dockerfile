FROM golang:alpine

RUN apk add socat

COPY ../ /root/slcan-svc

WORKDIR /root/slcan-svc

RUN go mod tidy \
    && go build .

# ENTRYPOINT [ "./scripts/entry.sh"]
