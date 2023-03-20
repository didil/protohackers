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
		err := s.processClientMsg(reqID, conn)
		if err != nil {
			s.speedDaemonSvc.UnregisterClient(reqID)
			if err == io.EOF {
				break
			}
			s.sendError(conn, err)
			break
		}

	}

}

func (s *Server) processClientMsg(reqID string, conn net.Conn) error {
	msgTypeData := make([]byte, 1)
	_, err := conn.Read(msgTypeData)
	if err != nil {
		return err
	}

	msgType, err := parseClientMessageType(msgTypeData)
	if err != nil {
		return err
	}

	switch msgType {
	case MsgTypeIAmCamera:
		iAmCamera, err := parseIAmCamera(conn)
		if err != nil {
			return err
		}

		err = s.speedDaemonSvc.RegisterAsCamera(reqID, int(iAmCamera.road), int(iAmCamera.mile), int(iAmCamera.limit))
		if err != nil {
			return err
		}
	case MsgTypeIAmDispatcher:
		iAmDispatcher, err := parseIAmDispatcher(conn)
		if err != nil {
			return err
		}

		err = s.speedDaemonSvc.RegisterAsDispatcher(reqID, iAmDispatcher.roads)
		if err != nil {
			return err
		}
	case MsgTypePlate:
		camera, err := s.speedDaemonSvc.GetCamera(reqID)
		if err != nil {
			// not registered as camera
			return err
		}

		plate, err := parsePlate(conn)
		if err != nil {
			return err
		}

		s.speedDaemonSvc.SavePlateObservation(plate.plate, plate.timestamp, camera.Road, camera.Mile, camera.Limit)

	}

	return nil
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
	road  int
	mile  int
	limit int
}

func parseIAmCamera(r io.Reader) (*IAmCamera, error) {
	buf := make([]byte, 6) // 2 + 2 + 2
	_, err := io.ReadFull(r, buf)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return nil, errors.New("message IAmCamera incomplete")
	}

	iAmCamera := &IAmCamera{}
	iAmCamera.road = int(binary.BigEndian.Uint16(buf[0:2]))
	iAmCamera.mile = int(binary.BigEndian.Uint16(buf[2:4]))
	iAmCamera.limit = int(binary.BigEndian.Uint16(buf[4:6]))

	return iAmCamera, nil
}

type IAmDispatcher struct {
	numRoads int
	roads    []int
}

func parseIAmDispatcher(r io.Reader) (*IAmDispatcher, error) {
	numRoadsBuf := make([]byte, 1)
	_, err := io.ReadFull(r, numRoadsBuf)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return nil, errors.New("message IAmDispatcher incomplete")
	}

	iAmDispatcher := &IAmDispatcher{}
	iAmDispatcher.numRoads = int(numRoadsBuf[0])

	roadsBuf := make([]byte, iAmDispatcher.numRoads*2)
	_, err = io.ReadFull(r, roadsBuf)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return nil, errors.New("message IAmDispatcher incomplete")
	}

	iAmDispatcher.roads = make([]int, iAmDispatcher.numRoads)

	for i := 0; i < iAmDispatcher.numRoads; i++ {
		iAmDispatcher.roads[i] = int(binary.BigEndian.Uint16(roadsBuf[2*i : 2*i+2]))
	}

	return iAmDispatcher, nil
}

type Plate struct {
	plate     string
	timestamp int
}

func parsePlate(r io.Reader) (*Plate, error) {
	plateLengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, plateLengthBuf)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return nil, errors.New("message Plate incomplete")
	}

	plateBuf := make([]byte, int(plateLengthBuf[0]))
	_, err = io.ReadFull(r, plateBuf)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return nil, errors.New("message Plate incomplete")
	}

	plate := &Plate{}
	plate.plate = string(plateBuf)

	timestampBuf := make([]byte, 4)
	_, err = io.ReadFull(r, timestampBuf)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return nil, errors.New("message Plate incomplete")
	}

	plate.timestamp = int(binary.BigEndian.Uint32(timestampBuf))

	return plate, nil
}

type WantHeartbeat struct {
	interval int
}

func parseWantHeartbeat(r io.Reader) (*WantHeartbeat, error) {
	wantHeartbeat := &WantHeartbeat{}

	intervalBuf := make([]byte, 4)
	_, err := io.ReadFull(r, intervalBuf)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return nil, errors.New("message Plate incomplete")
	}

	wantHeartbeat.interval = int(binary.BigEndian.Uint32(intervalBuf))

	return wantHeartbeat, nil
}
