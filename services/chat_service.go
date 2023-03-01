package services

import (
	"regexp"
	"sort"
	"sync"

	"go.uber.org/zap"
)

type ChatService interface {
	IsValidName(name string) bool
	AddUser(name string) (int, chan string)
	RemoveUser(id int)
	ListCurrentUsersNames() []string
	Broadcast(userId int, event string)
}

type chatService struct {
	nameRegex    *regexp.Regexp
	lastId       int
	users        map[int]*ChatUser
	userChannels map[int](chan string)
	lock         *sync.Mutex
	logger       *zap.Logger
}

type ChatUser struct {
	ID   int
	Name string
}

func NewChatService(logger *zap.Logger) ChatService {
	nameRegex := regexp.MustCompile("^[a-zA-Z0-9]*$")
	return &chatService{
		nameRegex:    nameRegex,
		lastId:       0,
		users:        map[int]*ChatUser{},
		userChannels: map[int](chan string){},
		lock:         &sync.Mutex{},
		logger:       logger,
	}
}

func (svc *chatService) IsValidName(name string) bool {
	if len(name) == 0 {
		return false
	}
	if len(name) > 16 {
		return false
	}

	return svc.nameRegex.MatchString(name)
}

const chatChannelsBuffer = 256

func (svc *chatService) AddUser(name string) (int, chan string) {
	svc.lock.Lock()
	defer svc.lock.Unlock()

	svc.lastId++

	user := &ChatUser{ID: svc.lastId, Name: name}
	c := make(chan string, chatChannelsBuffer)

	svc.users[user.ID] = user
	svc.userChannels[user.ID] = c

	return user.ID, c
}

func (svc *chatService) RemoveUser(id int) {
	svc.lock.Lock()
	defer svc.lock.Unlock()

	if _, ok := svc.users[id]; !ok {
		// user already removed
		return
	}

	delete(svc.users, id)

	channel := svc.userChannels[id]
	close(channel)
	delete(svc.userChannels, id)
}

func (svc *chatService) ListCurrentUsersNames() []string {
	names := make([]string, 0, len(svc.users))

	for _, u := range svc.users {
		names = append(names, u.Name)
	}

	sort.Strings(names)

	return names
}

func (svc *chatService) Broadcast(userId int, msg string) {
	svc.lock.Lock()
	defer svc.lock.Unlock()

	user := svc.users[userId]
	if user == nil {
		svc.logger.Error("user not found for announcement", zap.Int("userId", userId))
		return
	}

	for k, c := range svc.userChannels {
		// announce to other users only
		if k != userId {
			c <- msg
		}
	}
}
