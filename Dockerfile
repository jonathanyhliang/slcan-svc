FROM golang:1.19

COPY . /root/slcan-svc

WORKDIR /root/slcan-svc

RUN go mod tidy \
    && go build .

EXPOSE 8080

CMD ["./slcan-svc", "-p /dev/ttyACM0"]
