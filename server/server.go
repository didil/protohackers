package server

import (
	"context"
	"fmt"
	"net"

	"github.com/didil/protohackers/services"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type ProtoHackersMode string

type Server struct {
	mode           ProtoHackersMode
	port           int
	logger         *zap.Logger
	chatSvc        services.ChatService
	unusualDbSvc   services.UnusualDbService
	speedDaemonSvc services.SpeedDaemonService
}

const (
	ProtoHackersModeEcho            = "echo"
	ProtoHackersModePrimeTime       = "prime"
	ProtoHackersModeMeans           = "means"
	ProtoHackersModeBudgetChat      = "budget_chat"
	ProtoHackersModeUnusualDatabase = "ud"
	ProtoHackersModeMobInTheMiddle  = "mob"
	ProtoHackersModeSpeedDaemon     = "speed-daemon"
)

var validModes = []ProtoHackersMode{
	ProtoHackersModeEcho,
	ProtoHackersModePrimeTime,
	ProtoHackersModeMeans,
	ProtoHackersModeBudgetChat,
	ProtoHackersModeUnusualDatabase,
	ProtoHackersModeMobInTheMiddle,
	ProtoHackersModeSpeedDaemon,
}

type ServerOpt func(*Server) *Server

func NewServer(mode string, port int, logger *zap.Logger, opts ...ServerOpt) (*Server, error) {
	if !isValidMode(mode) {
		return nil, fmt.Errorf("invalid mode %s", mode)
	}
	s := &Server{
		mode:   ProtoHackersMode(mode),
		port:   port,
		logger: logger,
	}

	for _, opt := range opts {
		s = opt(s)
	}

	return s, nil
}

func WithChatService(chatSvc services.ChatService) ServerOpt {
	return func(s *Server) *Server {
		s.chatSvc = chatSvc
		return s
	}
}

func WithUnusualDbService(unusualDbSvc services.UnusualDbService) ServerOpt {
	return func(s *Server) *Server {
		s.unusualDbSvc = unusualDbSvc
		return s
	}
}

func WithSpeedDaemonDbService(speedDaemonSvc services.SpeedDaemonService) ServerOpt {
	return func(s *Server) *Server {
		s.speedDaemonSvc = speedDaemonSvc
		return s
	}
}

func isValidMode(mode string) bool {
	return slices.Contains(validModes, ProtoHackersMode(mode))
}

func (s *Server) Start(done <-chan bool) error {

	if s.mode == ProtoHackersModeUnusualDatabase {
		return s.StartUnusualDatabase(done)
	} else {
		return s.StartTCP(done)
	}

	return nil
}

func (s *Server) StartTCP(done <-chan bool) error {
	addr := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		return errors.Wrapf(err, "failed to start listener")
	}

	s.logger.Sugar().Infof("TCP Server listening on %s / mode: %s ...", addr, s.mode)

	go s.AcceptTCP(listener)

	<-done
	s.logger.Sugar().Infof("TCP: Received 'done' signal, closing listener")
	listener.Close()

	return nil
}

func (s *Server) AcceptTCP(listener net.Listener) {
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

		go s.HandleTCPConn(conn)
	}
}

type ContextKey string

var reqIDContextKey ContextKey = "req-id"

func (s *Server) HandleTCPConn(conn net.Conn) {
	ctx := context.Background()
	reqID := uuid.New().String()
	ctx = context.WithValue(ctx, reqIDContextKey, reqID)

	s.logger.Info("received connection", zap.String("reqID", reqID), zap.String("remote", conn.RemoteAddr().String()))

	switch s.mode {
	case ProtoHackersModeEcho:
		s.HandleEcho(ctx, conn)
	case ProtoHackersModePrimeTime:
		s.HandlePrimeTime(ctx, conn)
	case ProtoHackersModeMeans:
		s.HandleMeans(ctx, conn)
	case ProtoHackersModeBudgetChat:
		s.HandleBudgetChat(ctx, conn)
	case ProtoHackersModeMobInTheMiddle:
		s.HandleMobInTheMiddle(ctx, conn)
	case ProtoHackersModeSpeedDaemon:
		s.HandleSpeedDaemon(ctx, conn)
	default:
		panic("invalid mode: " + s.mode)
	}

	s.logger.Info("ended connection", zap.String("reqID", reqID), zap.String("remote", conn.RemoteAddr().String()))
}

func (s *Server) StartUnusualDatabase(done <-chan bool) error {
	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: s.port})
	if err != nil {
		return errors.Wrapf(err, "failed to start udp listener")
	}

	s.logger.Sugar().Infof("UDP Server listening on %d / mode: %s ...", s.port, s.mode)

	go s.HandleUnusualDatabase(udpConn)

	<-done
	s.logger.Sugar().Infof("UDP: Received 'done' signal, closing listener")
	udpConn.Close()

	return nil
}
