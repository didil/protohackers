package services

import (
	"errors"
	"math"
	"sync"

	"golang.org/x/exp/slices"
)

type SpeedDaemonService interface {
	RegisterAsCamera(reqID string, road int, mile int, limit int) error
	RegisterAsDispatcher(reqID string, roads []int) ([]chan *Ticket, error)
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
	// tickets: plate numbers that received tickets indexed by day
	tickets map[int][]string
	// plate observations indexed by plate number
	plates map[string][]*Plate
	// channel transmitting plate observations
	platesC chan *Plate
	// channels transmitting tickets, indexed by road number
	ticketsChannels map[int]chan *Ticket
	lock            *sync.Mutex
}

func NewSpeedDaemonService() SpeedDaemonService {
	s := &speedDaemonService{
		clients:         map[string]*Client{},
		cameras:         map[string]*Camera{},
		dispatchers:     map[string]*Dispatcher{},
		roads:           map[int][]string{},
		tickets:         map[int][]string{},
		plates:          map[string][]*Plate{},
		platesC:         make(chan *Plate, 1024),
		ticketsChannels: map[int](chan *Ticket){},
		lock:            &sync.Mutex{},
	}

	go s.processPlateObservations()

	return s
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

func (s *speedDaemonService) RegisterAsDispatcher(reqID string, roads []int) ([]chan *Ticket, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	client := s.clients[reqID]
	if client != nil {
		return nil, errors.New("client already registered")
	}

	s.clients[reqID] = &Client{
		clientType: ClientTypeDispatcher,
	}

	s.dispatchers[reqID] = &Dispatcher{
		roads: roads,
	}

	myTicketChans := []chan *Ticket{}

	for _, road := range roads {
		s.roads[road] = append(s.roads[road], reqID)

		if s.ticketsChannels[road] == nil {
			s.ticketsChannels[road] = make(chan *Ticket, 1024)
		}

		myTicketChans = append(myTicketChans, s.ticketsChannels[road])
	}

	return myTicketChans, nil
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
	for p := range s.platesC {
		s.lock.Lock()
		for _, pl := range s.plates[p.plateNumber] {
			if p.road == pl.road && pl.timestamp != p.timestamp {
				var p1, p2 *Plate
				if p.timestamp < pl.timestamp {
					p1 = p
					p2 = pl
				} else {
					p1 = pl
					p2 = p
				}

				speed := calcSpeed(p1, p2)
				if speed >= float64(p1.limit)+0.5 {
					s.issueTickets(p1, p2)
				}
			}
		}

		s.lock.Unlock()
	}
}

type Ticket struct {
	Plate      string
	Road       int
	Mile1      int
	Timestamp1 int
	Mile2      int
	Timestamp2 int
	Speed      int
}

func (s *speedDaemonService) issueTickets(p1, p2 *Plate) {
	// for multiday, we issue a single ticket, and we add to tickets index for each of the days
	days := getDays(p1.timestamp, p2.timestamp)

	for _, d := range days {
		if !slices.Contains(s.tickets[d], p1.plateNumber) {
			s.tickets[d] = append(s.tickets[d], p1.plateNumber)
		}
	}

	t := &Ticket{
		Plate:      p1.plateNumber,
		Road:       p1.road,
		Mile1:      p1.mile,
		Timestamp1: p1.timestamp,
		Mile2:      p2.mile,
		Timestamp2: p2.timestamp,
		Speed:      int(calcSpeed(p1, p2) * 100),
	}

	if s.ticketsChannels[p1.road] == nil {
		s.ticketsChannels[p1.road] = make(chan *Ticket, 1024)
	}

	s.ticketsChannels[p1.road] <- t
}

// speed in mph
func calcSpeed(p1, p2 *Plate) float64 {
	return (float64(p2.mile) - float64(p1.mile)) * 3600 / (float64(p2.timestamp) - float64(p1.timestamp))
}

// get days (start of day) between two timestamps
func getDays(start, end int) []int {
	days := []int{}

	if start > end {
		return days
	}

	startDay := int(math.Floor(float64(start) / 86400))
	endDay := int(math.Floor(float64(end) / 86400))

	for i := startDay; i <= endDay; i++ {
		days = append(days, i)
	}

	return days
}
