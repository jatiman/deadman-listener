package deadman

import (
	"testing"
	"time"

	"github.com/go-kit/log"
)

const (
	testDuration          = 100 * time.Millisecond
	shortIntervalDuration = 20 * time.Millisecond
	longIntervalDuration  = 30 * time.Millisecond
)

func TestDeadManDoesntTrigger(t *testing.T) {
	t.Helper()

	pinger := time.NewTicker(shortIntervalDuration)
	defer pinger.Stop()

	called := false

	logger := log.NewNopLogger()
	d := newDeadMan(pinger.C, shortIntervalDuration, func() error {
		called = true
		return nil
	}, logger)

	go d.Run()
	t.Cleanup(func() {
		d.Stop()
	})

	time.Sleep(testDuration)
	if called {
		t.Error("deadman triggered unexpectedly")
	}
}

func TestDeadManTriggers(t *testing.T) {
	t.Helper()

	pinger := time.NewTicker(longIntervalDuration)
	defer pinger.Stop()

	called := false

	logger := log.NewNopLogger()
	d := newDeadMan(pinger.C, shortIntervalDuration, func() error {
		called = true
		return nil
	}, logger)

	go d.Run()
	t.Cleanup(func() {
		d.Stop()
	})

	time.Sleep(testDuration)
	if !called {
		t.Error("deadman did not trigger as expected")
	}
}
