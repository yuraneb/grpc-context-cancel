package transceiver

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/yuraneb/grpc_context_cancel/api/grpc/pingpong"
)

type TxRx struct {
	rootCtx           context.Context
	terminate         func()
	id                int
	rxPort            int
	txPort            int
	rxDelay           time.Duration
	txTimeout         time.Duration
	reportIntervalSec int
	txAttempts        int
	// server     pingpong.PingServiceServer
	client pingpong.PingServiceClient
	wg     sync.WaitGroup
}

func NewTxRx(root context.Context, instance_id int, tx int, rx int, cTimeout time.Duration, serverDelay time.Duration, numAttempts int, cancelF func(), reportInterval int) *TxRx {
	return &TxRx{rootCtx: root, id: instance_id, rxPort: rx, txPort: tx, txTimeout: cTimeout, rxDelay: serverDelay,
		txAttempts: numAttempts, terminate: cancelF, reportIntervalSec: reportInterval}
}

// TxInit intializes the grpc client
func (t *TxRx) TxInit() {

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf(":%d", t.txPort), grpc.WithInsecure())
	if err != nil {
		logrus.Fatalf("Unable to establish Tx conn: %v", err)
	}
	t.wg.Add(1)

	go func() {

		<-t.rootCtx.Done()
		conn.Close()
		t.wg.Done()

	}()

	t.client = pingpong.NewPingServiceClient(conn)
}

// RxInit initializes the grpc server
func (t *TxRx) RxInit(relay bool) {

	servConn, err := net.Listen("tcp", fmt.Sprintf(":%d", t.rxPort))
	if err != nil {
		logrus.Fatalf("Unable to listen on server port: %v", err)
	}

	serverLog := logrus.New().WithField("serverID", t.id).WithField("service", "server")

	s := pingpong.Server{Logger: serverLog, Delay: t.rxDelay, Relay: relay, Client: t.client}

	grpcServer := grpc.NewServer()

	pingpong.RegisterPingServiceServer(grpcServer, &s)

	go func() {
		if err := grpcServer.Serve(servConn); err != nil {
			logrus.Fatalf("failed to serve: %s", err)
		}
	}()

	serverLog.Infof("Server listening on port %d", t.rxPort)
	t.wg.Add(1)

	// stat reporting thread
	go func() {

		statsTicker := time.NewTicker(time.Duration(time.Second * time.Duration(t.reportIntervalSec)))

		for {
			select {
			case <-t.rootCtx.Done():
				return
			case <-statsTicker.C:
				serverLog.Infof("Total messages received: %d", s.TotalMsg)
			}
		}

	}()

	go func() {

		<-t.rootCtx.Done()
		grpcServer.GracefulStop()
		t.wg.Done()

	}()
}

// RunTests runs a series of Tx test calls against the configured server
// Kicking off a series of cascaded GRPC calls
func (t *TxRx) RunTests() {

	clientLog := logrus.New().WithField("service", "client")

	for msg := 1; msg <= t.txAttempts; msg++ {

		clientCtx, cancelFunc := context.WithCancel(context.Background())

		// cancel/invalidate the context after specified delay
		// these goroutines can "temporarily leak" - exist past the usefulness of the context, but this has no effect on the test
		go func() {
			<-time.After(t.txTimeout)
			cancelFunc()
		}()

		clientLog.Infof("sending message %d to next service", msg)
		_, err := t.client.PingNext(clientCtx, &pingpong.Message{Id: fmt.Sprintf("%d", msg)})
		if err != nil {
			clientLog.Errorf("Error while sending: %v", err)
		}
	}

	clientLog.Infof("Attempted to send %d messages", t.txAttempts)
}

func (t *TxRx) WaitForTermination() {
	t.wg.Wait()
}

func (t *TxRx) Terminate() {
	t.terminate()
}
