package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"

	"go.uber.org/zap"
)

var chatMessageLimit = 1000

func (s *Server) HandleBudgetChat(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	reqID, _ := ctx.Value(reqIDContextKey).(string)

	welcomeMsg := "Welcome to budgetchat! What shall I call you?"

	_, err := conn.Write([]byte(welcomeMsg + "\n"))
	if err != nil {
		s.logger.Error("HandleBudgetChat write error", zap.Error(err))
		return
	}

	s.logger.Info("Welcome given", zap.String("reqID", reqID))

	sc := bufio.NewScanner(conn)

	ok := sc.Scan()
	if !ok {
		err := sc.Err()
		if err != nil {
			s.logger.Error("HandleBudgetChat name scan error", zap.Error(err))
			return
		}
	}

	name := string(sc.Bytes())
	if !s.chatSvc.IsValidName(name) {
		s.logger.Error("HandleBudgetChat invalid chat name", zap.String("name", name))
		return
	}

	s.logger.Info("New user name received", zap.String("name", name))

	// tell user about current room users
	userNames := s.chatSvc.ListCurrentUsersNames()
	currentUsersMsg := "* The room contains: " + strings.Join(userNames, ", ")

	_, err = conn.Write([]byte(currentUsersMsg + "\n"))
	if err != nil {
		s.logger.Error("HandleBudgetChat room contains write error", zap.Error(err))
		return
	}

	// add user to room
	userId, userChan := s.chatSvc.AddUser(name)
	s.logger.Info("New user added received", zap.String("name", name), zap.Int("userId", userId))

	go func() {
		for msg := range userChan {
			_, err = conn.Write([]byte(msg + "\n"))
			if err != nil {
				s.logger.Error("HandleBudgetChat write message error", zap.Error(err), zap.Int("userId", userId))
				s.removeChatUserAndAnnounce(userId, name)
				return
			}
		}
	}()

	// announce user joined to current users
	s.chatSvc.Broadcast(userId, fmt.Sprintf("* %s has entered the room", name))

	for sc.Scan() {
		data := sc.Bytes()
		if len(data) > chatMessageLimit {
			data = data[:chatMessageLimit]
		}
		s.chatSvc.Broadcast(userId, fmt.Sprintf("[%s] %s", name, string(data)))
	}

	err = sc.Err()
	if err != nil {
		s.logger.Error("HandleBudgetChat scan error", zap.Error(err))
		s.removeChatUserAndAnnounce(userId, name)
		return
	}

	s.removeChatUserAndAnnounce(userId, name)
}

func (s *Server) removeChatUserAndAnnounce(userId int, name string) {
	s.chatSvc.RemoveUser(userId)

	// announce user left to current users
	s.chatSvc.Broadcast(userId, fmt.Sprintf("* %s has left the room", name))
}
