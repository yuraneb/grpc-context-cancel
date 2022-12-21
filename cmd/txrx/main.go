package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yuraneb/grpc_context_cancel/transceiver"
)

type config struct {
	ServiceID         *int
	RXPort            *int
	TXPort            *int
	TXAttempts        *int
	OpMode            *string
	TxTimeoutMs       *int
	RxDelayMs         *int
	ReportIntervalSec *int
}

var svcConfig config

func init() {

	svcConfig.OpMode = flag.String("opmode", "relay", "mode of operation of service instance [client,relay,termination]")
	svcConfig.ServiceID = flag.Int("id", 1, "id of service instance")
	svcConfig.RXPort = flag.Int("rxport", 9000, "server port")
	svcConfig.TXPort = flag.Int("txport", 9000, "target server port")
	svcConfig.TXAttempts = flag.Int("attempts", 5, "number of attempts")
	svcConfig.TxTimeoutMs = flag.Int("txtimeout", 0, "time (in ms) before the grpc client cancels the request context")
	svcConfig.RxDelayMs = flag.Int("rxtimeout", 0, "simulated time (in ms) server takes to process a call")

	svcConfig.ReportIntervalSec = flag.Int("reportInterval", 5, "stats reporting interval (in seconds)")

}

func main() {

	ctx, cancel := context.WithCancel(context.Background())

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-sigc
		cancel()
	}()

	os.Exit(runWithReturnCode(ctx))
}

func runWithReturnCode(rootCtx context.Context) int {

	flag.Parse()

	transCtx, cancel := context.WithCancel(rootCtx)

	trans := transceiver.NewTxRx(transCtx, *svcConfig.ServiceID, *svcConfig.TXPort, *svcConfig.RXPort,
		time.Duration(*svcConfig.TxTimeoutMs)*time.Millisecond, time.Duration(*svcConfig.RxDelayMs)*time.Millisecond, *svcConfig.TXAttempts, cancel, *svcConfig.ReportIntervalSec)

	switch *svcConfig.OpMode {
	case "client":

		logrus.Infof("starting in client mode")

		trans.TxInit()
		trans.RunTests()
		trans.Terminate()
		trans.WaitForTermination()

		return 0

	case "relay":

		logrus.Infof("starting in relay mode")

		trans.TxInit()
		trans.RxInit(true)
		trans.WaitForTermination()

		return 0

	case "termination":

		logrus.Infof("starting in termination mode")

		trans.RxInit(false)
		trans.WaitForTermination()
		return 0

	default:
		logrus.Fatalf("invalid operation mode %s", &svcConfig.OpMode)
	}

	return 0

}
