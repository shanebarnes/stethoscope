package net

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNif_diff_Create(t *testing.T) {
	mapA := make(map[string]net.Interface)
	nextPtr := &mapA
	mapB := make(map[string]net.Interface)
	prevPtr := &mapB

	var allEvents, createEvents int
	err := diff(&prevPtr, &nextPtr, func(name string, ev WatchEvent, err error) error {
		allEvents++

		switch ev {
		case EventCreate:
			createEvents++
		}

		return nil
	})

	assert.Nil(t, err)
	assert.Greater(t, allEvents, 0)
	assert.Equal(t, allEvents, createEvents)
}

func TestNif_diff_Delete(t *testing.T) {
	mapA := make(map[string]net.Interface)
	nextPtr := &mapA
	mapB := make(map[string]net.Interface)
	prevPtr := &mapB

	diff(&prevPtr, &nextPtr, func(name string, ev WatchEvent, err error) error {
		return nil
	})

	// Simulate deletion by adding a dummy interface
	dummy := "dummyif4321"
	(*prevPtr)[dummy] = net.Interface{}

	var allEvents, deleteEvents int
	var eventName string
	err := diff(&prevPtr, &nextPtr, func(name string, ev WatchEvent, err error) error {
		allEvents++

		switch ev {
		case EventDelete:
			eventName = name
			deleteEvents++
		}

		return nil
	})

	assert.Nil(t, err)
	assert.Equal(t, allEvents, 1)
	assert.Equal(t, allEvents, deleteEvents)
	assert.Equal(t, dummy, eventName)
}

func TestNif_diff_Modify(t *testing.T) {
	mapA := make(map[string]net.Interface)
	nextPtr := &mapA
	mapB := make(map[string]net.Interface)
	prevPtr := &mapB

	diff(&prevPtr, &nextPtr, func(name string, ev WatchEvent, err error) error {
		return nil
	})

	var dummy string
	for key := range (*prevPtr) {
		dummy = key
		break
	}

	// Simulate modification by changing details of an existing interface
	nif := (*prevPtr)[dummy]
	nif.MTU = nif.MTU - 1
	(*prevPtr)[dummy] = nif

	var allEvents, modifyEvents int
	var eventName string
	err := diff(&prevPtr, &nextPtr, func(name string, ev WatchEvent, err error) error {
		allEvents++

		switch ev {
		case EventModify:
			eventName = name
			modifyEvents++
		}

		return nil
	})

	assert.Nil(t, err)
	assert.Equal(t, allEvents, 1)
	assert.Equal(t, allEvents, modifyEvents)
	assert.Equal(t, dummy, eventName)
}

func TestNif_Walk_CallbackError(t *testing.T) {
	var cbCount int = 0

	err := Walk(func(nif net.Interface, err error) error {
		cbCount++
		return syscall.EINTR
	})

	assert.Equal(t, 1, cbCount)
	assert.Equal(t, syscall.EINTR, err)
}

func TestNif_Walk_Complete(t *testing.T) {
	var cbCount int = 0

	err := Walk(func(nif net.Interface, err error) error {
		cbCount++
		return nil
	})

	assert.Greater(t, cbCount, 0)
	assert.Nil(t, err)
}

func TestNif_Walk_NoCallback(t *testing.T) {
	assert.Equal(t, syscall.EINVAL, Walk(nil))
}

func TestNif_Watch_CallbackError(t *testing.T) {
	var cbCount int = 0

	err := Watch(context.Background(), time.Second, func(name string, ev WatchEvent, err error) error {
		cbCount++
		return syscall.EINTR
	})

	assert.Equal(t, 1, cbCount)
	assert.Equal(t, syscall.EINTR, err)
}

func TestNif_Watch_Cancel(t *testing.T) {
	var cbCount int = 0

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Second)
		cancel()
	}()

	err := Watch(ctx, 100 * time.Millisecond, func(name string, ev WatchEvent, err error) error {
		cbCount++
		return nil
	})

	assert.Greater(t, cbCount, 0)
	assert.Equal(t, syscall.ECANCELED, err)
}

func TestNif_Watch_NoCallback(t *testing.T) {
	assert.Equal(t, syscall.EINVAL, Watch(context.Background(), time.Second, nil))
}

func TestNif_Watch_NoContext(t *testing.T) {
	assert.Equal(t, syscall.EINVAL, Watch(nil, time.Second, func(name string, ev WatchEvent, err error) error {
		return nil
	}))
}

func TestNif_Watch_NonPositiveTick(t *testing.T) {
	assert.Equal(t, syscall.EINVAL, Watch(context.Background(), 0 * time.Second, func(name string, ev WatchEvent, err error) error {
		return nil
	}))
}

func TestNif_WatchEventString_InvalidEvent(t *testing.T) {
	assert.Equal(t, "Invalid", WatchEventString(10))
	assert.Equal(t, "Invalid", WatchEventString(100))
	assert.Equal(t, "Invalid", WatchEventString(1000))
}

func TestNif_WatchEventString_ValidEvent(t *testing.T) {
	assert.Equal(t, "Null", WatchEventString(EventNull))
	assert.Equal(t, "Create", WatchEventString(EventCreate))
	assert.Equal(t, "Modify", WatchEventString(EventModify))
	assert.Equal(t, "Delete", WatchEventString(EventDelete))
}
