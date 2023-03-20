package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRegisterAsCamera(t *testing.T) {
	s := NewSpeedDaemonService()

	reqID1 := uuid.New().String()

	err := s.RegisterAsCamera(reqID1, 66, 100, 60)
	assert.NoError(t, err)

	err = s.RegisterAsCamera(reqID1, 75, 90, 50)
	assert.ErrorContains(t, err, "client already registered")

	camera, err := s.GetCamera(reqID1)
	assert.NoError(t, err)

	assert.Equal(t, &Camera{Road: 66, Mile: 100, Limit: 60}, camera)

	s.UnregisterClient(reqID1)

	_, err = s.GetCamera(reqID1)
	assert.ErrorContains(t, err, "client not registered")
}

func TestRegisterAsDispatcher(t *testing.T) {
	s := NewSpeedDaemonService()

	reqID1 := uuid.New().String()
	reqID2 := uuid.New().String()

	err := s.RegisterAsDispatcher(reqID1, []int{10, 25, 32})
	assert.NoError(t, err)

	err = s.RegisterAsDispatcher(reqID2, []int{15, 18, 25})
	assert.NoError(t, err)

	err = s.RegisterAsDispatcher(reqID1, []int{11, 26, 33})
	assert.ErrorContains(t, err, "client already registered")

	assert.Equal(t, []string{reqID1, reqID2}, s.GetReqIdsForRoad(25))

	dispatcher, err := s.GetDispatcher(reqID1)
	assert.NoError(t, err)

	assert.Equal(t, &Dispatcher{roads: []int{10, 25, 32}}, dispatcher)

	s.UnregisterClient(reqID1)

	_, err = s.GetDispatcher(reqID1)
	assert.ErrorContains(t, err, "client not registered")

	assert.Equal(t, []string{reqID2}, s.GetReqIdsForRoad(25))
}
