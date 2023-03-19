package server

import (
	"context"
	"encoding/binary"
	"io"
	"net"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (s *Server) HandleSpeedDaemon(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	reqID, _ := ctx.Value(reqIDContextKey).(string)

	for {
		msgTypeData := make([]byte, 1)
		_, err := conn.Read(msgTypeData)
		if err == io.EOF {
			return
		}

		msgType, err := parseClientMessageType(msgTypeData)
		if err != nil {
			s.sendError(conn, err)
			break
		}

		switch msgType {
		case MsgTypeIAmCamera:
			iAmCamera, err := s.parseIAmCamera(conn)
			if err != nil {
				s.sendError(conn, err)
				break
			}

			err = s.speedDaemonSvc.RegisterAsCamera(reqID, int(iAmCamera.road), int(iAmCamera.mile), int(iAmCamera.limit))
			if err != nil {
				s.sendError(conn, err)
				break
			}

		}

	}

}

func (s *Server) sendError(conn net.Conn, err error) {
	msg := err.Error()
	bufLen := 1 + 1 + len(msg) // 1 byte to store msg type + 1 byte to store str length
	buf := make([]byte, bufLen)
	buf[0] = byte(MsgTypeServerError)
	writeStringToBuf(buf, 1, msg)

	_, err = conn.Write(buf)
	if err != nil {
		s.logger.Error("failed to write to conn", zap.Error(err))
	}
}

func writeStringToBuf(buf []byte, i int, msg string) {
	buf[i] = byte(len(msg))
	copy(buf[i+1:], []byte(msg))
}

type MsgType byte

const (
	// Server Messages
	MsgTypeServerError MsgType = 0x10
	MsgTypeTicket      MsgType = 0x21
	MsgTypeHeartbeat   MsgType = 0x41

	// Client Messages
	MsgTypePlate         MsgType = 0x20
	MsgTypeWantHeartbeat MsgType = 0x40
	MsgTypeIAmCamera     MsgType = 0x80
	MsgTypeIAmDispatcher MsgType = 0x81
)

func parseClientMessageType(msgTypeData []byte) (MsgType, error) {
	switch msgTypeData[0] {
	case byte(MsgTypePlate):
		return MsgTypePlate, nil
	case byte(MsgTypeWantHeartbeat):
		return MsgTypeWantHeartbeat, nil
	case byte(MsgTypeIAmCamera):
		return MsgTypeIAmCamera, nil
	case byte(MsgTypeIAmDispatcher):
		return MsgTypeIAmDispatcher, nil

	default:
		return 0, errors.New("illegal msg")
	}
}

type IAmCamera struct {
	road  uint16
	mile  uint16
	limit uint16
}

func (s *Server) parseIAmCamera(r io.Reader) (*IAmCamera, error) {
	buf := make([]byte, 6) // 2 + 2 + 2
	_, err := io.ReadFull(r, buf)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return nil, errors.New("message IAmCamera incomplete")
	}

	iAmCamera := &IAmCamera{}
	iAmCamera.road = binary.BigEndian.Uint16(buf[0:2])
	iAmCamera.mile = binary.BigEndian.Uint16(buf[2:4])
	iAmCamera.limit = binary.BigEndian.Uint16(buf[4:6])

	return iAmCamera, nil
}
