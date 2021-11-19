package discordgo

import (
	"os"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

//////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////// VARS NEEDED FOR TESTING
var (
	dgGlobal    *Session // Stores a global discordgo user session
	dgBotGlobal *Session // Stores a global discordgo bot session

	envToken    = os.Getenv("DGU_TOKEN")  // Token to use when authenticating the user account
	envBotToken = os.Getenv("DGB_TOKEN")  // Token to use when authenticating the bot account
	envGuild    = os.Getenv("DG_GUILD")   // Guild ID to use for tests
	envChannel  = os.Getenv("DG_CHANNEL") // Channel ID to use for tests
	envAdmin    = os.Getenv("DG_ADMIN")   // User ID of admin user to use for tests
)

// SkipIfShort will call t.Skip() if go test was passed `-short`
func SkipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping test due to -short")
	}
}

// SkipOrGetSession will call t.Skip() if go test was passed `-short`, otherwise it returns
// the global Session object (which is created, if it doesn't already exist)
func SkipOrGetSession(t *testing.T) *Session {
	t.Helper()
	SkipIfShort(t)

	if dgGlobal == nil {
		var err error
		dgGlobal, err = New(envToken)
		if err != nil || dgGlobal == nil {
			t.Fatalf("session failed to initialize: %v", err)
		}
	}

	return dgGlobal
}

// SkipOrGetBotSession will call t.Skip() if go test was passed `-short`, otherwise it returns
// the global bot Session object (which is created, if it doesn't already exist)
func SkipOrGetBotSession(t *testing.T) *Session {
	t.Helper()
	SkipIfShort(t)

	// TODO: should this just fail if envBotToken isn't set?
	if dgBotGlobal == nil && envBotToken != "" {
		var err error
		dgBotGlobal, err = New(envBotToken)
		if err != nil || dgBotGlobal == nil {
			t.Fatalf("bot session failed to initialize: %v", err)
		}
	}

	return dgBotGlobal
}

//////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////// START OF TESTS

// TestNew tests the New() function without any arguments.  This should return
// a valid Session{} struct and no errors.
func TestNew(t *testing.T) {
	SkipIfShort(t)

	_, err := New()
	if err != nil {
		t.Errorf("New() returned error: %+v", err)
	}
}

// TestInvalidToken tests the New() function with an invalid token
func TestInvalidToken(t *testing.T) {
	SkipIfShort(t)

	d, err := New("asjkldhflkjasdh")
	if err != nil {
		t.Fatalf("New(InvalidToken) returned error: %+v", err)
	}

	// New with just a token does not do any communication, so attempt an api call.
	_, err = d.UserSettings()
	if err == nil {
		t.Errorf("New(InvalidToken), d.UserSettings returned nil error.")
	}
}

// TestNewToken tests the New() function with a Token.
func TestNewToken(t *testing.T) {
	SkipIfShort(t)

	if envToken == "" {
		t.Skip("Skipping New(token), DGU_TOKEN not set")
	}

	d, err := New(envToken)
	if err != nil {
		t.Fatalf("New(envToken) returned error: %+v", err)
	}

	if d == nil {
		t.Fatal("New(envToken), d is nil, should be Session{}")
	}

	if d.Token == "" {
		t.Fatal("New(envToken), d.Token is empty, should be a valid Token.")
	}
}

func TestOpenClose(t *testing.T) {
	SkipIfShort(t)
	if envToken == "" {
		t.Skip("Skipping TestClose, DGU_TOKEN not set")
	}

	d, err := New(envToken)
	if err != nil {
		t.Fatalf("TestClose, New(envToken) returned error: %+v", err)
	}

	if err = d.Open(); err != nil {
		t.Fatalf("TestClose, d.Open failed: %+v", err)
	}

	// We need a better way to know the session is ready for use,
	// this is totally gross.
	start := time.Now()
	for {
		d.RLock()
		if d.DataReady {
			d.RUnlock()
			break
		}
		d.RUnlock()

		if time.Since(start) > 10*time.Second {
			t.Fatal("DataReady never became true.yy")
		}
		runtime.Gosched()
	}

	// TODO find a better way
	// Add a small sleep here to make sure heartbeat and other events
	// have enough time to get fired.  Need a way to actually check
	// those events.
	time.Sleep(2 * time.Second)

	// UpdateStatus - maybe we move this into wsapi_test.go but the websocket
	// created here is needed.  This helps tests that the websocket was setup
	// and it is working.
	if err = d.UpdateGameStatus(0, time.Now().String()); err != nil {
		t.Errorf("UpdateStatus error: %+v", err)
	}

	if err = d.Close(); err != nil {
		t.Fatalf("TestClose, d.Close failed: %+v", err)
	}
}

func TestAddHandler(t *testing.T) {
	SkipIfShort(t)

	testHandlerCalled := int32(0)
	testHandler := func(s *Session, m *MessageCreate) {
		atomic.AddInt32(&testHandlerCalled, 1)
	}

	interfaceHandlerCalled := int32(0)
	interfaceHandler := func(s *Session, i interface{}) {
		atomic.AddInt32(&interfaceHandlerCalled, 1)
	}

	bogusHandlerCalled := int32(0)
	bogusHandler := func(s *Session, se *Session) {
		atomic.AddInt32(&bogusHandlerCalled, 1)
	}

	d := Session{}
	d.AddHandler(testHandler)
	d.AddHandler(testHandler)

	d.AddHandler(interfaceHandler)
	d.AddHandler(bogusHandler)

	d.handleEvent(messageCreateEventType, &MessageCreate{})
	d.handleEvent(messageDeleteEventType, &MessageDelete{})

	<-time.After(500 * time.Millisecond)

	// testHandler will be called twice because it was added twice.
	if atomic.LoadInt32(&testHandlerCalled) != 2 {
		t.Fatalf("testHandler was not called twice.")
	}

	// interfaceHandler will be called twice, once for each event.
	if atomic.LoadInt32(&interfaceHandlerCalled) != 2 {
		t.Fatalf("interfaceHandler was not called twice.")
	}

	if atomic.LoadInt32(&bogusHandlerCalled) != 0 {
		t.Fatalf("bogusHandler was called.")
	}
}

func TestRemoveHandler(t *testing.T) {
	SkipIfShort(t)

	testHandlerCalled := int32(0)
	testHandler := func(s *Session, m *MessageCreate) {
		atomic.AddInt32(&testHandlerCalled, 1)
	}

	d := Session{}
	r := d.AddHandler(testHandler)

	d.handleEvent(messageCreateEventType, &MessageCreate{})

	r()

	d.handleEvent(messageCreateEventType, &MessageCreate{})

	<-time.After(500 * time.Millisecond)

	// testHandler will be called once, as it was removed in between calls.
	if atomic.LoadInt32(&testHandlerCalled) != 1 {
		t.Fatalf("testHandler was not called once.")
	}
}
