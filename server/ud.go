package server

import (
	"bytes"
	"fmt"
	"net"

	"go.uber.org/zap"
)

var maxUDContentSize = 1000

func (s *Server) HandleUnusualDatabase(conn *net.UDPConn) {
	defer conn.Close()

	s.logger.Info("ud waiting for conns", zap.String("addr", conn.LocalAddr().String()))

	for {
		inputData := make([]byte, maxUDContentSize)
		n, addr, err := conn.ReadFromUDP(inputData)
		if err != nil {
			s.logger.Error("failed to read from udp", zap.Error(err))
			continue
		}

		s.logger.Info("received command", zap.String("command", string(inputData[:n])), zap.String("addr", addr.String()))

		go s.unusualDatabaseResponse(conn, addr, inputData[:n])
	}
}

func (s *Server) unusualDatabaseResponse(conn *net.UDPConn, addr *net.UDPAddr, inputData []byte) {
	if n := bytes.Index(inputData, []byte("=")); n > -1 {
		// set
		key := string(inputData[:n])
		var value string
		if len(inputData) == n+1 {
			value = ""
		} else {
			value = string(inputData[n+1:])
		}
		s.unusualDbSvc.Set(key, value)
	} else {
		// get
		key := string(inputData)
		value := s.unusualDbSvc.Get(key)

		_, err := conn.WriteToUDP([]byte(fmt.Sprintf("%s=%s", key, value)), addr)
		if err != nil {
			s.logger.Error("failed to write to udp", zap.Error(err))
			return
		}
	}
}
