FROM golang:1.19

COPY . /root/slcan-svc

WORKDIR /root/slcan-svc

RUN go mod tidy \
    && go build .
