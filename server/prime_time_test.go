package server

import (
	"bufio"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHandlePrimeTime(t *testing.T) {
	mode := ProtoHackersModePrimeTime
	port := 35001
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

	time.Sleep(500 * time.Millisecond)

	conn, err := net.DialTCP("tcp4", nil, tcpAddr)
	assert.NoError(t, err)
	defer conn.Close()

	sc := bufio.NewScanner(conn)

	// case 1

	writeData := []byte("{\"method\":\"isPrime\",\"number\":123}\n")
	_, err = conn.Write(writeData)
	assert.NoError(t, err)

	sc.Scan()

	readData := sc.Bytes()
	assert.NoError(t, sc.Err())

	assert.Equal(t, "{\"method\":\"isPrime\",\"prime\":false}", string(readData))

	// case 2

	writeData = []byte("{\"method\":\"isPrime\",\"number\":3}\n")
	_, err = conn.Write(writeData)
	assert.NoError(t, err)

	sc.Scan()

	readData = sc.Bytes()
	assert.NoError(t, sc.Err())

	assert.Equal(t, "{\"method\":\"isPrime\",\"prime\":true}", string(readData))

	// case 3

	writeData = []byte("{\"method\":\"isPrime\",\"number\":1.5}\n")
	_, err = conn.Write(writeData)
	assert.NoError(t, err)

	sc.Scan()

	readData = sc.Bytes()
	assert.NoError(t, sc.Err())

	assert.Equal(t, "{\"method\":\"isPrime\",\"prime\":false}", string(readData))

	// case 4

	writeData = []byte("{\"method\":\"isPrime\",\"numsber\":3}\n")
	_, err = conn.Write(writeData)
	assert.NoError(t, err)

	sc.Scan()

	readData = sc.Bytes()
	assert.NoError(t, sc.Err())

	assert.Equal(t, "ERROR", string(readData))
	done <- true
}

func TestHandlePrimeTimeMalfo(t *testing.T) {
	mode := ProtoHackersModePrimeTime
	port := 35002
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

	time.Sleep(500 * time.Millisecond)

	conn, err := net.DialTCP("tcp4", nil, tcpAddr)
	assert.NoError(t, err)
	defer conn.Close()

	sc := bufio.NewScanner(conn)

	// case 1

	writeData := []byte("Is 1700461 prime?\n")
	_, err = conn.Write(writeData)
	assert.NoError(t, err)

	sc.Scan()

	readData := sc.Bytes()
	assert.NoError(t, sc.Err())

	assert.Equal(t, "ERROR", string(readData))
	done <- true
}

func TestHandlePrimeTimeNotNum(t *testing.T) {
	mode := ProtoHackersModePrimeTime
	port := 35003
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

	time.Sleep(500 * time.Millisecond)

	conn, err := net.DialTCP("tcp4", nil, tcpAddr)
	assert.NoError(t, err)
	defer conn.Close()

	sc := bufio.NewScanner(conn)

	// case 1

	writeData := []byte("{\"method\":\"isPrime\",\"number\": 1.5.6}\n")
	_, err = conn.Write(writeData)
	assert.NoError(t, err)

	sc.Scan()

	readData := sc.Bytes()
	assert.NoError(t, sc.Err())

	assert.Equal(t, "ERROR", string(readData))
	done <- true
}

func TestHandlePrimeTimeStringNotNum(t *testing.T) {
	mode := ProtoHackersModePrimeTime
	port := 35004
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

	time.Sleep(500 * time.Millisecond)

	conn, err := net.DialTCP("tcp4", nil, tcpAddr)
	assert.NoError(t, err)
	defer conn.Close()

	sc := bufio.NewScanner(conn)

	// case 1

	writeData := []byte("{\"method\":\"isPrime\",\"number\": \"13\"}\n")
	_, err = conn.Write(writeData)
	assert.NoError(t, err)

	sc.Scan()

	readData := sc.Bytes()
	assert.NoError(t, sc.Err())

	assert.Equal(t, "ERROR", string(readData))
	done <- true
}
