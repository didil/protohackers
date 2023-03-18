package server

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHandleMobInTheMiddle(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	os.Setenv("MOB_UPSTREAM_HOST", "127.0.0.1")

	s, err := NewServer(ProtoHackersModeMobInTheMiddle, 12345, logger)
	assert.NoError(t, err)

	done := make(chan bool, 1)

	upstreamListener, err := net.Listen("tcp4", fmt.Sprintf(":%d", mobUpstreamPort))
	assert.NoError(t, err)

	msgReceivedByMobServer := ""
	go func() {
		buff := make([]byte, mobMsgLimit)
		conn, err := upstreamListener.Accept()
		assert.NoError(t, err)
		defer conn.Close()

		n, err := conn.Read(buff)
		assert.NoError(t, err)

		msgReceivedByMobServer = string(buff[:n])
	}()

	go func() {
		err := s.Start(done)
		assert.NoError(t, err)
	}()

	time.Sleep(100 * time.Millisecond)

	tcpConn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345})
	assert.NoError(t, err)
	defer tcpConn.Close()

	_, err = tcpConn.Write([]byte("[PinkCoder342] Please send the payment of 750 Boguscoins to 7oSd9NnEJYZh6QQHc0PhsEjWFegMTx\n"))
	assert.NoError(t, err)

	err = tcpConn.CloseWrite()
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, "[PinkCoder342] Please send the payment of 750 Boguscoins to 7YWHMfk9JZe0LM0g1ZauHuiSxhI\n", msgReceivedByMobServer)

	done <- true
	time.Sleep(100 * time.Millisecond)

}

func TestReplaceWithBogusCoin(t *testing.T) {
	assert.Equal(t, "Hi alice, please send payment to 7YWHMfk9JZe0LM0g1ZauHuiSxhI", replaceWithBogusCoin("Hi alice, please send payment to 7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX"))
	assert.Equal(t, "Hi alice, please send payment to 7YWHMfk9JZe0LM0g1ZauHuiSxhI ok ?", replaceWithBogusCoin("Hi alice, please send payment to 7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX ok ?"))
	assert.Equal(t, "7YWHMfk9JZe0LM0g1ZauHuiSxhI ok ?", replaceWithBogusCoin("7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX ok ?"))
}
