package main

import (
	"fmt"
	log "github.com/aclisp/log4go"
	"net/http"
)

type Monitor struct {
}

// StartPprof start http monitor.
func InitMonitor(binds []string) {
	m := new(Monitor)
	monitorServeMux := http.NewServeMux()
	monitorServeMux.HandleFunc("/monitor/ping", m.Ping)
	for _, addr := range binds {
		go func(bind string) {
			if err := http.ListenAndServe(bind, monitorServeMux); err != nil {
				log.Error("http.ListenAndServe(\"%s\", pprofServeMux) error(%v)", addr, err)
				panic(err)
			}
		}(addr)
	}
}

// monitor ping
func (m *Monitor) Ping(w http.ResponseWriter, r *http.Request) {
	for _, c := range routerServiceMap {
		if err := c.Available(); err != nil {
			http.Error(w, fmt.Sprintf("ping rpc error(%v)", err), http.StatusInternalServerError)
			return
		}
	}
	w.Write([]byte("ok"))
}
