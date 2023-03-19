package services

import (
	"errors"
	"sync"
)

type SpeedDaemonService interface {
	RegisterAsCamera(reqID string, road int, mile int, limit int) error
	GetCamera(reqID string) (*Camera, error)
}

type speedDaemonService struct {
	clients map[string]*Client
	cameras map[string]*Camera
	lock    *sync.Mutex
}

func NewSpeedDaemonService() SpeedDaemonService {
	return &speedDaemonService{
		clients: map[string]*Client{},
		cameras: map[string]*Camera{},
		lock:    &sync.Mutex{},
	}
}

type ClientType int

const (
	ClientTypeCamera     ClientType = 1
	ClientTypeDispatcher ClientType = 2
)

type Client struct {
	clientType ClientType
}

type Camera struct {
	road  int
	mile  int
	limit int
}

func (s *speedDaemonService) RegisterAsCamera(reqID string, road int, mile int, limit int) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	client := s.clients[reqID]
	if client != nil {
		return errors.New("client already registered")
	}

	s.clients[reqID] = &Client{
		clientType: ClientTypeCamera,
	}

	s.cameras[reqID] = &Camera{
		road:  road,
		mile:  mile,
		limit: limit,
	}

	return nil
}

func (s *speedDaemonService) GetCamera(reqID string) (*Camera, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	client := s.clients[reqID]
	if client == nil {
		return nil, errors.New("client not registered")
	}
	if client.clientType != ClientTypeCamera {
		return nil, errors.New("client is not a camera")
	}

	camera := s.cameras[reqID]
	if camera == nil {
		return nil, errors.New("camera not found")
	}

	return camera, nil
}
