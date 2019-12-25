package net

import (
	"context"
	"net"
	"reflect"
	"syscall"
	"time"
)

type WalkFunc func(net.Interface, error) error

type WatchEvent uint
const (
	EventNull   WatchEvent = iota
	EventCreate
	EventModify
	EventDelete
)

type WatchFunc func(string, WatchEvent, error) error
var eventText = map[WatchEvent]string{
	EventNull:   "Null",
	EventCreate: "Create",
	EventModify: "Modify",
	EventDelete: "Delete",
}

func diff(prevPtr, nextPtr **map[string]net.Interface, fn WatchFunc) error {
	err := Walk(func(nif net.Interface, err error) error {
		ev := EventNull
		(**nextPtr)[nif.Name] = nif

		if _, ok := (**prevPtr)[nif.Name]; ok { // Detect modify event
			if !reflect.DeepEqual((**prevPtr)[nif.Name], nif) {
				ev = EventModify
			}
			delete(**prevPtr, nif.Name)
		} else { // Detect create event
			ev = EventCreate
		}

		if ev == EventNull {
			return nil
		} else {
			return fn(nif.Name, ev, err)
		}
	})

	// Detect delete event(s)
	if len(**prevPtr) > 0 && err == nil {
		for key := range **prevPtr {
			delete(**prevPtr, key)
			if err = fn(key, EventDelete, nil); err != nil {
				break
			}
		}
	}

	tmpPtr := *nextPtr
	*nextPtr = *prevPtr
	*prevPtr = tmpPtr

	return err
}

func WatchEventString(ev WatchEvent) string {
	str, ok := eventText[ev]
	if !ok {
		str = "Invalid"
	}
	return str
}

func Walk(fn WalkFunc) error {
	var err error

	if fn == nil {
		err = syscall.EINVAL
	} else {
		var ifs []net.Interface
		if ifs, err = net.Interfaces(); err == nil {
			for _, nif := range ifs {
				if err = fn(nif, nil); err != nil {
					break
				}
			}
		}
	}

	return err
}

func Watch(ctx context.Context, tick time.Duration, fn WatchFunc) error {
	var err error

	if ctx == nil || tick <= 0 * time.Second || fn == nil {
		err = syscall.EINVAL
	} else {
		ticker := time.NewTicker(tick)
		defer ticker.Stop()

		// Use two maps to poll for A/B diffs
		mapA := make(map[string]net.Interface)
		nextPtr := &mapA
		mapB := make(map[string]net.Interface)
		prevPtr := &mapB

		// Force first tick without waiting for tick interval
		select {
		case <-ctx.Done():
			err = syscall.ECANCELED
		default:
			err = diff(&prevPtr, &nextPtr, fn)
		}

		// Poll loop
		for err == nil {
			select {
			case <-ctx.Done():
				err = syscall.ECANCELED // alternatively: ctx.Err()
			case <-ticker.C:
				err = diff(&prevPtr, &nextPtr, fn)
			}
		}
	}

	return err
}
