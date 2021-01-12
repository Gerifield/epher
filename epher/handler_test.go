package epher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

type testRoom struct {
	broadcastError error
	broadcastMessage []byte
}

func (tr *testRoom) AddUser(u *User) {
	panic("implement me")
}

func (tr *testRoom) DelUser(u *User) {
	panic("implement me")
}

func (tr *testRoom) UserCount() int {
	panic("implement me")
}

func (tr *testRoom) BroadcastText(b []byte) error {
	tr.broadcastMessage = b
	return tr.broadcastError
}

func TestPushHandlerOK(t *testing.T) {
	e := New()
	testRoom := &testRoom{}
	e.Rooms["test1"] = testRoom

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/publish/test1", strings.NewReader("something"))

	// Chi router context magic
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("room", "test1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	e.PushHandler(rr, req)
	assert.Nil(t, testRoom.broadcastError)
	assert.Equal(t, "something", string(testRoom.broadcastMessage))
}
