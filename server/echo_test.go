package server

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHandleEcho(t *testing.T) {
	mode := ProtoHackersModeEcho
	port := 35000
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	s, err := NewServer(mode, port, logger, nil)
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

	testString := "ABCDEFG"

	writeData := []byte(testString)
	_, err = conn.Write(writeData)
	assert.NoError(t, err)

	err = conn.CloseWrite()
	assert.NoError(t, err)

	readData, err := io.ReadAll(conn)
	if err == io.EOF {
		err = nil
	}
	assert.NoError(t, err)

	assert.Equal(t, testString, string(readData))
	done <- true
	time.Sleep(100 * time.Millisecond)
}
