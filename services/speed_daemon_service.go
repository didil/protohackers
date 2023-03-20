package services

import (
	"errors"
	"sync"
)

type SpeedDaemonService interface {
	RegisterAsCamera(reqID string, road int, mile int, limit int) error
	RegisterAsDispatcher(reqID string, roads []int) error
	GetCamera(reqID string) (*Camera, error)
	GetDispatcher(reqID string) (*Dispatcher, error)
	UnregisterClient(reqId string)
	GetReqIdsForRoad(road int) []string
	SavePlateObservation(plate string, timestamp, road, mile, limit int)
}

type speedDaemonService struct {
	// client indexed by req id
	clients map[string]*Client
	// cameras indexed by req id
	cameras map[string]*Camera
	// dispatchers indexed by req id
	dispatchers map[string]*Dispatcher
	// req ids indexed by road number
	roads map[int][]string
	// tickets indexed by day
	tickets map[int][]string
	// plate observations indexed by plate number
	plates map[string][]*Plate
	// channel transmitting plate observations
	platesC chan *Plate
	lock    *sync.Mutex
}

func NewSpeedDaemonService() SpeedDaemonService {
	return &speedDaemonService{
		clients:     map[string]*Client{},
		cameras:     map[string]*Camera{},
		dispatchers: map[string]*Dispatcher{},
		roads:       map[int][]string{},
		tickets:     map[int][]string{},
		plates:      map[string][]*Plate{},
		platesC:     make(chan *Plate, 100),
		lock:        &sync.Mutex{},
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
	Road  int
	Mile  int
	Limit int
}

type Dispatcher struct {
	roads []int
}

type Plate struct {
	plateNumber string
	timestamp   int
	road        int
	mile        int
	limit       int
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
		Road:  road,
		Mile:  mile,
		Limit: limit,
	}

	return nil
}

func (s *speedDaemonService) RegisterAsDispatcher(reqID string, roads []int) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	client := s.clients[reqID]
	if client != nil {
		return errors.New("client already registered")
	}

	s.clients[reqID] = &Client{
		clientType: ClientTypeDispatcher,
	}

	s.dispatchers[reqID] = &Dispatcher{
		roads: roads,
	}

	for _, road := range roads {
		s.roads[road] = append(s.roads[road], reqID)
	}

	return nil
}

func (s *speedDaemonService) UnregisterClient(reqId string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.clients, reqId)
	delete(s.cameras, reqId)
	delete(s.dispatchers, reqId)

	for roadIdx, reqIds := range s.roads {
		s.roads[roadIdx] = deleteElemFromSlice(reqIds, reqId)
	}
}

func deleteElemFromSlice(s []string, elem string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == elem {
			s[i] = s[len(s)-1] // Copy last element to index i.
			s[len(s)-1] = ""   // Erase last element (write zero value).
			s = s[:len(s)-1]   // Truncate slice.
			break
		}
	}

	return s
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

func (s *speedDaemonService) GetDispatcher(reqID string) (*Dispatcher, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	client := s.clients[reqID]
	if client == nil {
		return nil, errors.New("client not registered")
	}
	if client.clientType != ClientTypeDispatcher {
		return nil, errors.New("client is not a dispatcher")
	}

	dispatcher := s.dispatchers[reqID]
	if dispatcher == nil {
		return nil, errors.New("dispatcher not found")
	}

	return dispatcher, nil
}

func (s *speedDaemonService) GetReqIdsForRoad(road int) []string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.roads[road]
}

func (s *speedDaemonService) SavePlateObservation(plateNumber string, timestamp, road, mile, limit int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	plate := &Plate{
		plateNumber,
		timestamp,
		road,
		mile,
		limit,
	}

	s.plates[plateNumber] = append(s.plates[plateNumber], plate)

	s.platesC <- plate
}

func (s *speedDaemonService) processPlateObservations() {

	for _ = range s.platesC {

		// ignore multiday for now

	}

}
