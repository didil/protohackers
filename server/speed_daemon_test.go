package server

import (
	"bytes"
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

func TestParseIAmCamera(t *testing.T) {
	buf, err := hex.DecodeString("00420064003c")
	assert.NoError(t, err)

	r := bytes.NewReader(buf)
	iAmCamera, err := parseIAmCamera(r)
	assert.NoError(t, err)

	assert.Equal(t, 66, int(iAmCamera.road))
	assert.Equal(t, 100, int(iAmCamera.mile))
	assert.Equal(t, 60, int(iAmCamera.limit))
}

func TestParseIAmCamera2(t *testing.T) {
	buf, err := hex.DecodeString("017004d20028")
	assert.NoError(t, err)

	r := bytes.NewReader(buf)
	iAmCamera, err := parseIAmCamera(r)
	assert.NoError(t, err)

	assert.Equal(t, 368, int(iAmCamera.road))
	assert.Equal(t, 1234, int(iAmCamera.mile))
	assert.Equal(t, 40, int(iAmCamera.limit))
}

func TestParseIAmDispatcher(t *testing.T) {
	buf, err := hex.DecodeString("010042")
	assert.NoError(t, err)

	r := bytes.NewReader(buf)
	iAmDispatcher, err := parseIAmDispatcher(r)
	assert.NoError(t, err)

	assert.Equal(t, 1, iAmDispatcher.numRoads)
	assert.Equal(t, []int{66}, iAmDispatcher.roads)
}

func TestParseIAmDispatcher2(t *testing.T) {
	buf, err := hex.DecodeString("03004201701388")
	assert.NoError(t, err)

	r := bytes.NewReader(buf)
	iAmDispatcher, err := parseIAmDispatcher(r)
	assert.NoError(t, err)

	assert.Equal(t, 3, iAmDispatcher.numRoads)
	assert.Equal(t, []int{66, 368, 5000}, iAmDispatcher.roads)
}

func TestParsePlate(t *testing.T) {
	buf, err := hex.DecodeString("04554e3158000003e8")
	assert.NoError(t, err)

	r := bytes.NewReader(buf)
	plate, err := parsePlate(r)
	assert.NoError(t, err)

	assert.Equal(t, "UN1X", plate.plate)
	assert.Equal(t, 1000, plate.timestamp)
}

func TestParsePlate2(t *testing.T) {
	buf, err := hex.DecodeString("0752453035424b470001e240")
	assert.NoError(t, err)

	r := bytes.NewReader(buf)
	plate, err := parsePlate(r)
	assert.NoError(t, err)

	assert.Equal(t, "RE05BKG", plate.plate)
	assert.Equal(t, 123456, plate.timestamp)
}

func TestParseWantHeartbeat(t *testing.T) {
	buf, err := hex.DecodeString("0000000a")
	assert.NoError(t, err)

	r := bytes.NewReader(buf)
	wantHeartbeat, err := parseWantHeartbeat(r)
	assert.NoError(t, err)

	assert.Equal(t, 10, wantHeartbeat.interval)
}

func TestParseWantHeartbeat2(t *testing.T) {
	buf, err := hex.DecodeString("000004db")
	assert.NoError(t, err)

	r := bytes.NewReader(buf)
	wantHeartbeat, err := parseWantHeartbeat(r)
	assert.NoError(t, err)

	assert.Equal(t, 1243, wantHeartbeat.interval)
}
