package server

import (
	"context"
	"io"
	"net"

	"go.uber.org/zap"
)

func (s *Server) HandleEcho(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	bufSize := 4096
	data := make([]byte, bufSize)

	var totalRead int
	var totalWritten int

	for {
		n, err := conn.Read(data)
		if err == io.EOF {
			break
		}
		if err != nil {
			s.logger.Error("HandleEcho read error", zap.Error(err))
			break
		}

		totalRead += n

		p, err := conn.Write(data[:n])
		if err != nil {
			s.logger.Error("HandleEcho write error", zap.Error(err))
			break
		}

		totalWritten += p
	}

	reqID, _ := ctx.Value(reqIDContextKey).(string)
	s.logger.Info("echo results",
		zap.String("reqID", reqID),
		zap.String("remote", conn.RemoteAddr().String()),
		zap.Int("totalRead", totalRead),
		zap.Int("totalWritten", totalWritten),
	)
}
