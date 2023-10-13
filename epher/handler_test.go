package epher

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

type testRoom struct {
	users            []*User
	broadcastError   error
	broadcastMessage []byte
}

func (tr *testRoom) AddUser(u *User) {
	tr.users = append(tr.users, u)
}

func (tr *testRoom) DelUser(u *User) {
	for i, v := range tr.users {
		if v.ID == u.ID {
			tr.users = append(tr.users[:i], tr.users[i+1:]...)
			break
		}
	}
}

func (tr *testRoom) UserCount() int {
	return len(tr.users)
}

func (tr *testRoom) BroadcastText(b []byte) error {
	tr.broadcastMessage = b
	return tr.broadcastError
}

func (tr *testRoom) Stop() {}

func TestPushHandlerOK(t *testing.T) {
	e := New(nil)
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

func TestPushHandlerMissingRoom(t *testing.T) {
	e := New(nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/publish/test1", strings.NewReader("something"))

	// Chi router context magic
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("room", "test1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	e.PushHandler(rr, req)
	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Equal(t, "no_room", rr.Body.String())
}

func TestAddConnectionWithoutRoom(t *testing.T) {
	e := New(nil)

	_, ok := e.Rooms["test1"]
	assert.False(t, ok)

	e.addConnection("test1", &User{ID: 1})
	_, ok = e.Rooms["test1"]
	assert.True(t, ok)
	assert.Equal(t, 1, e.Rooms["test1"].UserCount())

	e.addConnection("test1", &User{ID: 2})
	assert.Equal(t, 2, e.Rooms["test1"].UserCount())
}

func TestAddConnectionWithRoom(t *testing.T) {
	e := New(nil)
	testRoom := &testRoom{}

	e.Rooms["test1"] = testRoom
	assert.Len(t, testRoom.users, 0)

	e.addConnection("test1", &User{ID: 1})
	assert.Len(t, testRoom.users, 1)
}

func TestDelConnectionDropRoom(t *testing.T) {
	e := New(nil)
	testRoom := &testRoom{}

	e.Rooms["test1"] = testRoom
	_, ok := e.Rooms["test1"] // Room should be there
	assert.True(t, ok)
	assert.Equal(t, 0, e.Rooms["test1"].UserCount()) // No user there

	e.delConnection("test1", &User{ID: 1}) // Trigger a delete with a non-existing user
	_, ok = e.Rooms["test1"]               // Room should be deleted
	assert.False(t, ok)
}

type testUpgrader struct {
	upgradError error
}

func (t testUpgrader) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error) {
	return nil, t.upgradError
}

func TestWebsocketHandlerUpgradeFail(t *testing.T) {
	e := New(nil)

	// Make sure we store then restore the global state
	oldUpgrader := upgrader
	defer func() { upgrader = oldUpgrader }()

	upgrader = testUpgrader{
		upgradError: errors.New("something"),
	}
	rr := httptest.NewRecorder()
	e.WebsocketHandler(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, "websocket_upgrade_failed\n", rr.Body.String())
}
