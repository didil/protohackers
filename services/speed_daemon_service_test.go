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

	_, err := s.RegisterAsDispatcher(reqID1, []int{10, 25, 32})
	assert.NoError(t, err)

	_, err = s.RegisterAsDispatcher(reqID2, []int{15, 18, 25})
	assert.NoError(t, err)

	_, err = s.RegisterAsDispatcher(reqID1, []int{11, 26, 33})
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

func TestCalcSpeed(t *testing.T) {
	p1 := &Plate{
		mile:      8,
		timestamp: 100,
	}

	p2 := &Plate{
		mile:      9,
		timestamp: 145,
	}

	assert.Equal(t, float64(80), calcSpeed(p1, p2))
}

func TestGetDays(t *testing.T) {
	t1 := 1679127100 // March 18, 2023 8:11:40 AM
	t2 := 1679377515 // March 21, 2023 5:45:15 AM
	t3 := 1679393500 // March 21, 2023 10:11:40 AM

	assert.Equal(t, []int{19434, 19435, 19436, 19437}, getDays(t1, t2))
	assert.Equal(t, []int{19437}, getDays(t2, t3))
}
