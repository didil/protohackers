package services

import (
	"sync"
)

type UnusualDbService interface {
	Set(key string, value string)
	Get(key string) string
}

type unusualDbService struct {
	db   map[string]string
	lock *sync.Mutex
}

func NewUnusualDbService() UnusualDbService {
	return &unusualDbService{
		db: map[string]string{
			versionKey: "Ken's Key-Value Store 1.0",
		},
		lock: &sync.Mutex{},
	}
}

var versionKey string = "version"

func (s *unusualDbService) Set(key string, value string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if key == versionKey {
		return
	}

	s.db[key] = value
}

func (s *unusualDbService) Get(key string) string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.db[key]
}
