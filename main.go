package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/log"
	"github.com/jonathanyhliang/slcan-svc/docs"
)

//	@title		Serial-Line CAN Service API
//	@version	1.0

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host	localhost:port/slcan

func main() {
	var (
		httpAddr = flag.String("a", ":8080", "HTTP listen address")
		amqpURL  = flag.String("u", "amqp://guest:guest@localhost:5672/", "AMQP dialing address")
		port     = flag.String("p", "", "SLCAN port")
		baud     = flag.Int("b", 115200, "SLCAN port baudrate")
	)
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var b Backend
	{
		b = NewSlcanBackend()
	}

	var s Service
	{
		s = NewSlcanService()
		s = BackendMiddleware(b)(s)
		s = LoggingMiddleware(logger)(s)
	}

	var h http.Handler
	{
		h = MakeHTTPHandler(s, log.With(logger, "component", "HTTP"))
	}

	errs := make(chan error)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		errs <- b.Handler(*port, *baud, *amqpURL)
	}()

	go func() {
		docs.SwaggerInfo.BasePath = "/"
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errs <- http.ListenAndServe(*httpAddr, h)
	}()

	logger.Log("exit", <-errs)
}
