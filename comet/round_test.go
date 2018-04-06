package main

import (
	"testing"
)

func TestRound(t *testing.T) {
	options := RoundOptions{}
	r := NewRound(options)
	t0 := r.Timer(0)
	if t0 == nil {
		t.FailNow()
	}
	t1 := r.Timer(1)
	if t1 == nil {
		t.FailNow()
	}
	t2 := r.Timer(2)
	if t2 == nil {
		t.FailNow()
	}
	if t0 != t2 {
		t.FailNow()
	}
}
