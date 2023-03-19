package server

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseClientMessageType(t *testing.T) {
	msgType, err := parseClientMessageType([]byte{byte(MsgTypeIAmCamera)})

	assert.NoError(t, err)
	assert.Equal(t, MsgTypeIAmCamera, msgType)

	msgType, err = parseClientMessageType([]byte{byte(MsgTypePlate)})

	assert.NoError(t, err)
	assert.Equal(t, MsgTypePlate, msgType)

	msgType, err = parseClientMessageType([]byte{byte(MsgTypeWantHeartbeat)})

	assert.NoError(t, err)
	assert.Equal(t, MsgTypeWantHeartbeat, msgType)

	msgType, err = parseClientMessageType([]byte{byte(MsgTypeIAmDispatcher)})

	assert.NoError(t, err)
	assert.Equal(t, MsgTypeIAmDispatcher, msgType)

	_, err = parseClientMessageType([]byte{byte(MsgTypeServerError)})
	assert.ErrorContains(t, err, "illegal msg")

	_, err = parseClientMessageType([]byte{byte(MsgTypeServerError)})
	assert.ErrorContains(t, err, "illegal msg")

	_, err = parseClientMessageType([]byte{byte(MsgTypeHeartbeat)})
	assert.ErrorContains(t, err, "illegal msg")
}

func TestWriteStringToBuf(t *testing.T) {
	msg1 := "foo"
	buf1 := make([]byte, 1+len(msg1))
	writeStringToBuf(buf1, 0, msg1)

	assert.Equal(t, "03666f6f", hex.EncodeToString(buf1))

	msg2 := "Elbereth"
	buf2 := make([]byte, 1+len(msg1)+1+len(msg2))
	copy(buf2, buf1)
	writeStringToBuf(buf2, 1+len(msg1), msg2)

	assert.Equal(t, "03666f6f08456c626572657468", hex.EncodeToString(buf2))
}
