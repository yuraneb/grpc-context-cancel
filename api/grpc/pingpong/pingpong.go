package pingpong

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

type Server struct {
	Logger   logrus.FieldLogger
	Delay    time.Duration
	Client   PingServiceClient
	Relay    bool
	TotalMsg int // total messages seen by this server/relay
}

func (s *Server) PingNext(ctx context.Context, msg *Message) (*Message, error) {

	logger := s.Logger.WithField("messageID", msg.Id).WithField("service", "server")
	logger.Info("Starting handler")

	if s.Relay {
		relayLogger := logger.WithField("service", "relay")
		relayLogger.Info("Relaying call to next service")
		_, err := s.Client.PingNext(ctx, msg)
		if err != nil {
			relayLogger.Errorf("Got error from server :%v", err)
		}
	}

	select {
	case <-time.After(s.Delay):
		logger.Infof("No cancelled context, returning out after %v ms", s.Delay.Milliseconds())

	case <-ctx.Done():
		logger.Infof("Context terminated: %v", ctx.Err())

	}

	s.TotalMsg++
	return &Message{Id: msg.Id}, nil
}

func (s *Server) mustEmbedUnimplementedPingServiceServer() {}
