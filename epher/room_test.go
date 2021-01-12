package epher

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicRoomFunctions(t *testing.T) {
	r := NewRoom("test1")
	testUser1 := &User{ID: 1}
	testUser2 := &User{ID: 2}

	assert.Equal(t, 0, r.UserCount())

	r.AddUser(testUser1)
	assert.Equal(t, 1, r.UserCount())
	r.AddUser(testUser2)
	assert.Equal(t, 2, r.UserCount())

	r.DelUser(testUser1)
	assert.Equal(t, 1, r.UserCount())

	r.DelUser(testUser2)
	assert.Equal(t, 0, r.UserCount())
}

type testTextSender struct {
	called int
}

func (tts *testTextSender) SendText(b []byte) error {
	tts.called++
	return nil
}

func TestBroadcastText(t *testing.T) {
	r := NewRoom("test1")
	testUser1 := &User{
		ID:       1,
		connLock: &sync.Mutex{},
	}
	testUser2 := &User{
		ID:       2,
		connLock: &sync.Mutex{},
	}

	tts := &testTextSender{}
	testUser1.TextSender = tts // Overwrite/implement the TextSender function here!
	testUser2.TextSender = tts

	r.AddUser(testUser1)
	r.AddUser(testUser2)

	err := r.BroadcastText([]byte("test message"))
	assert.NoError(t, err)
	assert.Equal(t, 2, tts.called)
}
