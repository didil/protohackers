package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestIsValidName(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)
	svc := NewChatService(logger)

	assert.Equal(t, false, svc.IsValidName(""))
	assert.Equal(t, false, svc.IsValidName("12345678901234567"))
	assert.Equal(t, false, svc.IsValidName("mike_lois"))
	assert.Equal(t, false, svc.IsValidName(" john "))
	assert.Equal(t, false, svc.IsValidName("john.com"))

	assert.Equal(t, true, svc.IsValidName("mike"))
	assert.Equal(t, true, svc.IsValidName("mikeM123"))
	assert.Equal(t, true, svc.IsValidName("1234567890123456"))
}

func TestAddUsersListCurrentUsersNames(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)
	svc := NewChatService(logger)

	id, _ := svc.AddUser("mike")
	assert.Equal(t, 1, id)
	assert.Equal(t, []string{"mike"}, svc.ListCurrentUsersNames())

	id, _ = svc.AddUser("john")
	assert.Equal(t, 2, id)
	assert.Equal(t, []string{"john", "mike"}, svc.ListCurrentUsersNames())

	id, _ = svc.AddUser("mike")
	assert.Equal(t, 3, id)
	assert.Equal(t, []string{"john", "mike", "mike"}, svc.ListCurrentUsersNames())
}

func TestAnnounceUser(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)
	svc := NewChatService(logger)
	_, c1 := svc.AddUser("mike")
	_, c2 := svc.AddUser("lara")

	id, c := svc.AddUser("john")

	broadcastMessage := "* john has entered the room"

	svc.Broadcast(id, broadcastMessage)

	msg1 := <-c1
	assert.Equal(t, broadcastMessage, msg1)
	msg2 := <-c2
	assert.Equal(t, broadcastMessage, msg2)

	// close channel to make sure it's empty
	close(c)
	msg := <-c
	assert.Equal(t, "", msg)
}
