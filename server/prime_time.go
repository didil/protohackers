package server

import (
	"bufio"
	"context"
	"encoding/json"
	"math"
	"net"

	"go.uber.org/zap"
)

type PrimeTimeRequest struct {
	Method *string      `json:"method"`
	Number *json.Number `json:"number"`
}

type PrimeTimeResponse struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func (s *Server) HandlePrimeTime(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	sc := bufio.NewScanner(conn)

	numRequests := 0

	for sc.Scan() {
		req := &PrimeTimeRequest{}
		data := sc.Bytes()
		s.logger.Info("HandlePrimeTime request", zap.ByteString("data", data))

		err := json.Unmarshal(data, req)
		if err != nil {
			s.logger.Error("HandlePrimeTime unmarshall error", zap.ByteString("data", data), zap.Error(err))
			conn.Write([]byte("ERROR"))
			return
		}

		if !isValidPrimeRequest(req) {
			s.logger.Error("HandlePrimeTime invalid prime request", zap.ByteString("data", data), zap.Error(err))
			conn.Write([]byte("ERROR"))
			return
		}

		resp := &PrimeTimeResponse{
			Method: "isPrime",
			Prime:  isPrime(*req.Number),
		}

		out, err := json.Marshal(resp)
		if err != nil {
			s.logger.Error("HandlePrimeTime marshal error", zap.Error(err))
			break
		}

		out = append(out, '\n')

		_, err = conn.Write(out)
		if err != nil {
			s.logger.Error("HandlePrimeTime write error", zap.Error(err))
			break
		}

		numRequests++
	}

	err := sc.Err()
	if err != nil {
		s.logger.Error("HandlePrimeTime scan error", zap.Error(err))
		return

	}

	reqID, _ := ctx.Value(reqIDContextKey).(string)
	s.logger.Info("echo results",
		zap.String("reqID", reqID),
		zap.String("remote", conn.RemoteAddr().String()),
		zap.Int("numRequests", numRequests),
	)
}

func isValidPrimeRequest(req *PrimeTimeRequest) bool {
	if req == nil {
		return false
	}
	if req.Method == nil {
		return false
	}
	if req.Number == nil {
		return false
	}

	_, errF := req.Number.Float64()
	_, errI := req.Number.Float64()
	if errF != nil && errI != nil {
		return false
	}

	if *req.Method != "isPrime" {
		return false
	}

	return true
}

func isPrime(number json.Number) bool {
	n, err := number.Int64()
	if err != nil {
		return false
	}

	if n <= 1 {
		return false
	}

	for i := 2; i <= int(math.Sqrt(float64(n))); i++ {
		if n%int64(i) == 0 {
			return false
		}
	}

	return true
}
