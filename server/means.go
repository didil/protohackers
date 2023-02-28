package server

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"go.uber.org/zap"
)

type PriceRequestType byte

const (
	PriceRequestTypeInsert PriceRequestType = 'I'
	PriceRequestTypeQuery  PriceRequestType = 'Q'
)

type PriceRequest struct {
	RequestType PriceRequestType
	A           int32
	B           int32
}

type PriceDBEntry struct {
	Timestamp int32
	Price     int32
}

func (s *Server) HandleMeans(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	bufSize := 9
	data := make([]byte, bufSize)

	db := []*PriceDBEntry{}

	var totalRead int
	var totalWritten int

	for {
		n, err := io.ReadFull(conn, data)
		if err == io.EOF {
			break
		}
		if err != nil {
			s.logger.Error("HandleMeans read error", zap.Error(err))
			break
		}

		totalRead += n

		req, err := parsePriceRequest(data)
		if err != nil {
			s.logger.Error("HandleMeans parsePriceRequest error", zap.Error(err))
			break
		}

		if req.RequestType == PriceRequestTypeInsert {
			db = insertPrice(db, &PriceDBEntry{Timestamp: req.A, Price: req.B})
		} else if req.RequestType == PriceRequestTypeQuery {
			avP := averagePrice(db, req.A, req.B)

			respData := make([]byte, 4)
			binary.BigEndian.PutUint32(respData, uint32(avP))

			p, err := conn.Write(respData)
			if err != nil {
				s.logger.Error("HandleMeans write error", zap.Error(err))
				break
			}

			totalWritten += p
		} else {
			s.logger.Error("HandleMeans unknown query type error", zap.String("requestType", string(req.RequestType)))
			break
		}

	}

	reqID, _ := ctx.Value(reqIDContextKey).(string)
	s.logger.Info("means results",
		zap.String("reqID", reqID),
		zap.String("remote", conn.RemoteAddr().String()),
		zap.Int("totalRead", totalRead),
		zap.Int("totalWritten", totalWritten),
	)
}

func parsePriceRequest(data []byte) (*PriceRequest, error) {
	req := &PriceRequest{}

	if len(data) != 9 {
		return nil, fmt.Errorf("data length should be 9")
	}

	reqTypeRaw := data[0]
	if reqTypeRaw == 'I' {
		req.RequestType = PriceRequestTypeInsert
	} else if reqTypeRaw == 'Q' {
		req.RequestType = PriceRequestTypeQuery
	} else {
		return nil, fmt.Errorf("unknown request type %q", reqTypeRaw)
	}

	req.A = int32(binary.BigEndian.Uint32(data[1:5]))
	req.B = int32(binary.BigEndian.Uint32(data[5:9]))

	return req, nil
}

func insertPrice(db []*PriceDBEntry, entry *PriceDBEntry) []*PriceDBEntry {
	db = append(db, entry)
	return db
}

func averagePrice(db []*PriceDBEntry, minT int32, maxT int32) int32 {
	var sum int64
	var count int64

	for i := 0; i < len(db); i++ {
		if db[i].Timestamp >= minT && db[i].Timestamp <= maxT {
			sum += int64(db[i].Price)
			count++
		}
	}

	if count == 0 {
		return 0
	}

	avg := sum / count

	return int32(avg)
}
