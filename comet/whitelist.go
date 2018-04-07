package main

import (
	"fmt"
	"strings"

	log "github.com/aclisp/log4go"
)

type Whitelist struct {
	Log  log.Logger
	list map[string]struct{} // whitelist for debug
}

// NewWhitelist a whitelist struct.
func NewWhitelist(file string, list []string) (w *Whitelist, err error) {
	w = new(Whitelist)
	w.Log = make(log.Logger)
	w.list = make(map[string]struct{})
	for _, key := range list {
		w.list[key] = struct{}{}
	}

	flw := log.NewFileLogWriter(file, false)
	if flw == nil {
		err = fmt.Errorf("can not open file: %s", file)
		return
	}
	flw.SetFormat("[%D %T] [%L] [%S] %M")
	flw.SetRotateSize(100 * 1024 * 1024)
	flw.SetRotate(true)
	w.Log.AddFilter("file", log.INFO, flw)
	return
}

// Contains whitelist contains a key or not.
func (w *Whitelist) Contains(key string) (ok bool) {
	if ix := strings.Index(key, "_"); ix > -1 {
		_, ok = w.list[key[:ix]]
	}
	return
}

func (w *Whitelist) Close() {
	w.Log.Close()
}
