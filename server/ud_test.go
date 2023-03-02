package server

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/didil/protohackers/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHandleUnusualDatabase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mode := ProtoHackersModeUnusualDatabase
	port := 35000
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	uDSvc := mocks.NewMockUnusualDbService(ctrl)

	key := "my-key"
	value := "my-value"

	uDSvc.EXPECT().Set(key, value).Return()
	uDSvc.EXPECT().Get(key).Return(value)

	s, err := NewServer(mode, port, logger, WithUnusualDbService(uDSvc))
	assert.NoError(t, err)

	done := make(chan bool, 1)

	go func() {
		err := s.Start(done)
		assert.NoError(t, err)
	}()

	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: port,
	}

	time.Sleep(100 * time.Millisecond)

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	assert.NoError(t, err)
	defer conn.Close()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		outputData := make([]byte, maxUDContentSize)
		n, _, err := conn.ReadFromUDP(outputData)
		assert.NoError(t, err)

		assert.Equal(t, "my-key=my-value", string(outputData[:n]))

		wg.Done()
	}()

	inputData := []byte(key + "=" + value)
	n, err := conn.Write(inputData)
	assert.NoError(t, err)

	logger.Info("wrote set query", zap.Int("n", n))

	inputData = []byte(key)
	_, err = conn.Write(inputData)
	assert.NoError(t, err)

	wg.Wait()
}
