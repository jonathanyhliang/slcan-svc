#!/bin/sh
socat pty,link=/dev/vmodem0,waitslave tcp:host.minikube.internal:80
./slcan-svc -p /dev/vmodem0
