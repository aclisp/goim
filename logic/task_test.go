package main

import (
	"fmt"
	"testing"
	"time"
)

func TestTaskWithGoRoutine(t *testing.T) {
	a := 1
	task := func() {
		a++
		fmt.Println("task done", a)
	}
	go task()
}

func TestTaskWithChannel(t *testing.T) {
	ch := make(chan func(), 10)
	go func() {
		v := <-ch
		v()
	}()

	a := 1
	task := func() {
		a++
		fmt.Println("task done", a)
	}
	ch <- task
	time.Sleep(1*time.Second)
}