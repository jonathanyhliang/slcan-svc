FROM golang:1.19

RUN apt update \
    && apt install -y --no-install-recommends socat 

COPY . /root/slcan-svc

WORKDIR /root/slcan-svc

RUN go mod tidy \
    && go build .
