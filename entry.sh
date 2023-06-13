#!/bin/sh
socat pty,link=/dev/vmodem0,waitslave tcp:127.0.0.1:80
./slcan-svc -p /dev/vmodem0
