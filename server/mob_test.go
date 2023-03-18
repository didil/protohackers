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
		assert.NotNil(t, upstreamListener)
		conn, err := upstreamListener.Accept()
		assert.NoError(t, err)
		defer conn.Close()

		buff := make([]byte, mobMsgLimit)
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
	assert.Equal(t, "Hi alice, please send payment to 7YWHMfk9JZe0LM0g1ZauHuiSxhI", replaceWithBogusCoin2("Hi alice, please send payment to 7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX"))
	assert.Equal(t, "Hi alice, please send payment to 7YWHMfk9JZe0LM0g1ZauHuiSxhI ok 7YWHMfk9JZe0LM0g1ZauHuiSxhI ?", replaceWithBogusCoin2("Hi alice, please send payment to 7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX ok 7mQ06fryM9E3IXQ1tR6RSNdIn9qcLwkxedp ?"))
	assert.Equal(t, "7YWHMfk9JZe0LM0g1ZauHuiSxhI ok ?", replaceWithBogusCoin2("7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX ok ?"))
	assert.Equal(t, "[TinyCharlie994] This is a product ID, not a Boguscoin: 76wHjKPI3t7zCZPSUJaN8Wu1uwoVAKCN-u4vsTdErgoL9PZviChc2Jp0iNXkWgo-1234", replaceWithBogusCoin2("[TinyCharlie994] This is a product ID, not a Boguscoin: 76wHjKPI3t7zCZPSUJaN8Wu1uwoVAKCN-u4vsTdErgoL9PZviChc2Jp0iNXkWgo-1234"))
	assert.Equal(t, "Hi alice, please send payment to 7YWHMfk9JZe0LM0g1ZauHuiSxhI", replaceWithBogusCoin2("Hi alice, please send payment to 7mQ06fryM9E3IXQ1tR6RSNdIn9qcLwkxedp"))
	assert.Equal(t, "[ProtoWizard91] Please pay the ticket price of 15 Boguscoins to one of these addresses: 7YWHMfk9JZe0LM0g1ZauHuiSxhI 7YWHMfk9JZe0LM0g1ZauHuiSxhI 7YWHMfk9JZe0LM0g1ZauHuiSxhI", replaceWithBogusCoin2("[ProtoWizard91] Please pay the ticket price of 15 Boguscoins to one of these addresses: 7YWHMfk9JZe0LM0g1ZauHuiSxhI 7mQ06fryM9E3IXQ1tR6RSNdIn9qcLwkxedp 7YWHMfk9JZe0LM0g1ZauHuiSxhI"))
}
