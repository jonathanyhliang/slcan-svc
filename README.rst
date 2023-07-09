Serial-Line CAN Microservice
############################

Overview
########

``slcan-svc`` is a `go-kit <https://github.com/go-kit/kit>`_ based microserivce impelementation
which is capable of bridging serial-line CAN communication via RESTful APIs. The serial-line
backend of the service is partially `Serial-Line CAN <https://github.com/torvalds/linux/blob/master/drivers/net/can/slcan/slcan-core.c>`_
compliant for interfacing a `slcan <https://github.com/jonathanyhliang/zephyr/tree/slcan/samples/subsys/canbus/slcan>`_ end device.

``slcan-svc`` could interact with `mcumgr-svc <https://github.com/jonathanyhliang/mcumgr-svc>`_ using ``RabbitMQ`` to perform
``Device Firmware Update``. See `demo-svc <https://github.com/jonathanyhliang/demo-svc>`_ for the detail.

Building and Running
####################

Clone and build ``slcan-svc`` repository:

.. code-block:: console

        cd /workdir/slcan-svc
        go build -o ./build/cli ./cli

Get **slcan-svc** usage:

.. code-block:: console

        ./workdir/build/cli -h

         Usage of /workdir/slcan-svc/build/cli:
        -a string
                HTTP listen address (default ":8080")
        -b int
                SLCAN port baudrate (default 115200)
        -p string
                SLCAN port
        -u string
                AMQP dialing address (default "amqp://guest:guest@localhost:5672/")


Run **slcan-svc** with serial port specified:

.. code-block:: console

       ./workdir/build/cli -p /dev/ttyACM0

To access the RESTful APIs:

.. code-block:: console

        curl http://localhost:8080/slcan \
                --include --header "Content-Type: application/json" \
                --request "POST" \
                --data "{"id": 123, "data": "200rpm"}"
        
        curl http://localhost:8080/slcan/123 \
                --include --header "Content-Type: application/json" \
                --request "GET"

        curl http://localhost:8080/slcan/123 \
                --include --header "Content-Type: application/json" \
                --request "PUT" \
                --data "{"id": 123, "data": "300rpm"}"
        
        curl http://localhost:8080/slcan/123 \
                --include --header "Content-Type: application/json" \
                --request "DELETE"

