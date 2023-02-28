package server

import (
	"encoding/binary"
	"encoding/hex"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestParsePriceRequest_Insert(t *testing.T) {
	src := []byte("490000303900000065")
	data := make([]byte, hex.DecodedLen(len(src)))
	_, err := hex.Decode(data, src)
	assert.NoError(t, err)

	req, err := parsePriceRequest(data)
	assert.NoError(t, err)

	assert.Equal(t, PriceRequestTypeInsert, req.RequestType)
	assert.Equal(t, int32(12345), req.A)
	assert.Equal(t, int32(101), req.B)
}

func TestParsePriceRequest_Query(t *testing.T) {
	src := []byte("51000003e8000186a0")
	data := make([]byte, hex.DecodedLen(len(src)))
	_, err := hex.Decode(data, src)
	assert.NoError(t, err)

	req, err := parsePriceRequest(data)
	assert.NoError(t, err)

	assert.Equal(t, PriceRequestTypeQuery, req.RequestType)
	assert.Equal(t, int32(1000), req.A)
	assert.Equal(t, int32(100000), req.B)
}

func TestHandleMeans(t *testing.T) {
	mode := ProtoHackersModeMeans
	port := 35000
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	s, err := NewServer(mode, port, logger)
	assert.NoError(t, err)

	done := make(chan bool, 1)

	go func() {
		err := s.Start(done)
		assert.NoError(t, err)
	}()

	tcpAddr := &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: port,
	}

	time.Sleep(100 * time.Millisecond)

	conn, err := net.DialTCP("tcp4", nil, tcpAddr)
	assert.NoError(t, err)
	defer conn.Close()

	writeHex(t, "490000303900000065", conn)
	writeHex(t, "490000303a00000066", conn)
	writeHex(t, "490000303b00000064", conn)
	writeHex(t, "490000a00000000005", conn)
	writeHex(t, "510000300000004000", conn)

	err = conn.CloseWrite()
	assert.NoError(t, err)

	readData, err := io.ReadAll(conn)
	if err == io.EOF {
		err = nil
	}
	assert.NoError(t, err)

	assert.Len(t, readData, 4)

	res := binary.BigEndian.Uint32(readData)

	assert.Equal(t, int32(101), int32(res))
	done <- true
	time.Sleep(100 * time.Millisecond)
}

func writeHex(t *testing.T, h string, conn *net.TCPConn) {
	src := []byte(h)
	writeData := make([]byte, hex.DecodedLen(len(src)))
	_, err := hex.Decode(writeData, src)
	assert.NoError(t, err)

	_, err = conn.Write(writeData)
	assert.NoError(t, err)
}
