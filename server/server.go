package server

import (
	"context"
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type ProtoHackersMode string

type Server struct {
	mode   ProtoHackersMode
	port   int
	logger *zap.Logger
}

const (
	ProtoHackersModeEcho = "echo"
)

var validModes = []ProtoHackersMode{
	ProtoHackersModeEcho,
}

func NewServer(mode string, port int, logger *zap.Logger) (*Server, error) {
	if !isValidMode(mode) {
		return nil, fmt.Errorf("invalid mode %s", mode)
	}
	s := &Server{
		mode:   ProtoHackersMode(mode),
		port:   port,
		logger: logger,
	}

	return s, nil
}

func isValidMode(mode string) bool {
	return slices.Contains(validModes, ProtoHackersMode(mode))
}

func (s *Server) Start(done <-chan bool) error {
	addr := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		return errors.Wrapf(err, "failed to start listener")
	}

	s.logger.Sugar().Infof("Server listening on %s ...", addr)

	go s.Accept(listener)

	<-done
	s.logger.Sugar().Infof("Received 'done' signal, closing listener")
	listener.Close()

	return nil
}

func (s *Server) Accept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			s.logger.Sugar().Infof("Accept: conn closed")
			return
		}
		if err != nil {
			s.logger.Sugar().Errorf("failed to accept conn: %v\n", err)
			return
		}

		go s.HandleConn(conn)
	}
}

type ContextKey string

var reqIDContextKey ContextKey = "req-id"

func (s *Server) HandleConn(conn net.Conn) {
	ctx := context.Background()
	reqID := uuid.New().String()
	ctx = context.WithValue(ctx, reqIDContextKey, reqID)

	s.logger.Info("received connection", zap.String("reqID", reqID), zap.String("remote", conn.RemoteAddr().String()))

	switch s.mode {
	case ProtoHackersModeEcho:
		s.HandleEcho(ctx, conn)
	default:
		panic("invalid mode: " + s.mode)
	}

	s.logger.Info("ended connection", zap.String("reqID", reqID), zap.String("remote", conn.RemoteAddr().String()))
}
