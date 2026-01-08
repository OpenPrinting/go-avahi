// CGo binding for Avahi
//
// Copyright (C) 2025 by Prashant Andoriya
// See LICENSE for license terms and conditions
//
// Unit tests for eventqueue behavior
//
//go:build linux || freebsd

package avahi

import (
	"testing"
	"time"
)

// TestEventQueueBasic verifies basic push and receive behavior
func TestEventQueueBasic(t *testing.T) {
	var q eventqueue[int]
	q.init()

	q.Push(1)
	q.Push(2)
	q.Push(3)

	ch := q.Chan()

	if v := <-ch; v != 1 {
		t.Fatalf("expected 1, got %d", v)
	}
	if v := <-ch; v != 2 {
		t.Fatalf("expected 2, got %d", v)
	}
	if v := <-ch; v != 3 {
		t.Fatalf("expected 3, got %d", v)
	}

	q.Close()
}

// TestEventQueueClose verifies that Close eventually closes the output channel
func TestEventQueueClose(t *testing.T) {
	var q eventqueue[int]
	q.init()

	q.Push(42)
	q.Close()

	_, ok := <-q.Chan()
	if ok {
		t.Fatalf("expected channel to be closed after Close()")
	}
}

func TestEventQueueCloseEventuallyCloses(t *testing.T) {
	var q eventqueue[int]
	q.init()

	q.Push(1)
	q.Push(2)
	q.Push(3)

	q.Close()

	select {
	case _, ok := <-q.Chan():
		if ok {
			_, ok = <-q.Chan()
			if ok {
				t.Fatalf("expected channel to be closed eventually after Close()")
			}
		}
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for channel to close")
	}
}

// TestEventQueueMultiplePush verifies multiple pushes before reading
func TestEventQueueMultiplePush(t *testing.T) {
	var q eventqueue[int]
	q.init()

	for i := 0; i < 10; i++ {
		q.Push(i)
	}

	for i := 0; i < 10; i++ {
		select {
		case v := <-q.Chan():
			if v != i {
				t.Fatalf("expected %d, got %d", i, v)
			}
		case <-time.After(time.Second):
			t.Fatalf("timeout waiting for value %d", i)
		}
	}

	q.Close()
}
