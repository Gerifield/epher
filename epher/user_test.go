package epher

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

type testWsConn struct {
	readMessageError error

	nextWriterMessageType int
	nextWriterStream      io.WriteCloser
	nextWriterError       error
}

func (twc *testWsConn) ReadMessage() (messageType int, p []byte, err error) {
	return 0, nil, twc.readMessageError
}

func (twc *testWsConn) NextWriter(messageType int) (io.WriteCloser, error) {
	twc.nextWriterMessageType = messageType
	return twc.nextWriterStream, twc.nextWriterError
}

func TestNewUser(t *testing.T) {
	u := NewUser(nil)
	assert.NotEqual(t, 0, u.ID)
	assert.NotNil(t, u.TextSender)
}

func TestReadLoopConnectionError(t *testing.T) {
	testError := errors.New("something went wrong")
	twc := &testWsConn{
		readMessageError: testError,
	}
	u := NewUser(twc)

	assert.Equal(t, testError, u.ReadLoop())
}

func TestReadLoopConnectionWSError(t *testing.T) {
	testError := &websocket.CloseError{Code: websocket.CloseAbnormalClosure}
	twc := &testWsConn{
		readMessageError: testError,
	}
	u := NewUser(twc)

	assert.Equal(t, testError, u.ReadLoop())
}

func TestReadLoopConnectionCleanClose(t *testing.T) {
	twc := &testWsConn{
		readMessageError: &websocket.CloseError{Code: websocket.CloseNormalClosure},
	}
	u := NewUser(twc)

	assert.NoError(t, u.ReadLoop())
}

func TestSendTextNextWriterFail(t *testing.T) {
	testError := errors.New("something went wrong")
	twc := &testWsConn{
		nextWriterError: testError,
	}
	u := NewUser(twc)

	assert.Equal(t, testError, u.SendText([]byte("test")))
	assert.Equal(t, websocket.TextMessage, twc.nextWriterMessageType) // It was called
}

type buffCloser struct {
	io.Writer
}

func (buffCloser) Close() error { return nil }

func TestSendTextOK(t *testing.T) {
	buff := &bytes.Buffer{}
	twc := &testWsConn{
		nextWriterStream: buffCloser{Writer: buff},
	}
	u := NewUser(twc)

	assert.NoError(t, u.SendText([]byte("test")))
	assert.Equal(t, websocket.TextMessage, twc.nextWriterMessageType)
	assert.Equal(t, "test", buff.String())
}
