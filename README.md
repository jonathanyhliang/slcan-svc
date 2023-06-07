# Serial-Line CAN Microservice


## Overview


This repository implements a [go-kit](https://github.com/go-kit/kit) based microserivce which is capable of bridging serial-line CAN communication via RESTful APIs. The serial-line backend of the service is a [Serial-Line CAN](https://github.com/torvalds/linux/blob/master/drivers/net/can/slcan/slcan-core.c) compliant end device.


## Building and Running


1. Clone **slcan-svc** repository
2. ``cd $workspace/slcan-svc``
3. ``go build .``
4. Get **slcan-svc** usage by: ``./slcan-svc -h``. Run **slcan-svc** with serial port specified, for example: ``./slcan-svc -p /dev/ttyACM0``


```
    Usage of $workspace/slcan-svc:
        -a string
                HTTP listen address (default ":8080")
        -b int
                SLCAN port baudrate (default 115200)
        -p string
                SLCAN port
        -u string
                AMQP dialing address (default "amqp://guest:guest@localhost:5672/")
```

5. Use curl to interact with RESTful APIs

```
curl http://localhost:8080/slcan --include --header "Content-Type: application/json" \
--request "POST" --data "{"id": 123, "data": "200rpm"}"
```

## Intergrating with MCUmgr Service


**Slcan-svc** could be integrated with [mcumgr-svc](https://github.com/jonathanyhliang/mcumgr-svc) to accomplish firmware update of slcan end deivce via serial port. [Mcumgr](https://github.com/apache/mynewt-mcumgr) is a golang library capable of managing MCU over a variety means of comminucation interfaces. In order to coordinate the services for firmware update, RabbitMQ is utilised to allow **slcan-svc** to ping **mcumgr-svc** for serial port handover.

![slcan-group](/plantuml/diag/slcan-group/slcan-group.png)


## API Documentation

 
![api docs](/plantuml/diag/swag_ui.png)
