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

	assert.Equal(t, &Camera{road: 66, mile: 100, limit: 60}, camera)
}
