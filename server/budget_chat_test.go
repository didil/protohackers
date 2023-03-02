package server

import (
	"bufio"
	"net"
	"testing"
	"time"

	"github.com/didil/protohackers/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHandleBudgetChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mode := ProtoHackersModeBudgetChat
	port := 35000
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	chatSvc := mocks.NewMockChatService(ctrl)

	myUserName := "peter"
	userId := 101
	userChan := make(chan string, 256)

	chatSvc.EXPECT().IsValidName(myUserName).Return(true)
	chatSvc.EXPECT().ListCurrentUsersNames().Return([]string{"danny", "eva"})
	chatSvc.EXPECT().AddUser(myUserName).Return(userId, userChan)
	chatSvc.EXPECT().Broadcast(userId, "* peter has entered the room").Return()
	chatSvc.EXPECT().Broadcast(userId, "[peter] Hello folks").Return()
	chatSvc.EXPECT().Broadcast(userId, "[peter] Bye folks").Return()
	chatSvc.EXPECT().RemoveUser(userId).Return()
	chatSvc.EXPECT().Broadcast(userId, "* peter has left the room").Return()

	s, err := NewServer(mode, port, logger, WithChatService(chatSvc))
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

	sc := bufio.NewScanner(conn)
	assert.True(t, sc.Scan())
	assert.NoError(t, sc.Err())

	welcomeMsg := string(sc.Bytes())
	assert.Equal(t, "Welcome to budgetchat! What shall I call you?", welcomeMsg)

	_, err = conn.Write([]byte(myUserName + "\n"))
	assert.NoError(t, err)

	assert.True(t, sc.Scan())
	assert.NoError(t, sc.Err())

	roomContainsMsg := string(sc.Bytes())
	assert.Equal(t, "* The room contains: danny, eva", roomContainsMsg)

	_, err = conn.Write([]byte("Hello folks" + "\n"))
	assert.NoError(t, err)

	_, err = conn.Write([]byte("Bye folks" + "\n"))
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	err = conn.Close()
	assert.NoError(t, err)

	done <- true
	time.Sleep(100 * time.Millisecond)
}
