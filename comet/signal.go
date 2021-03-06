package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/aclisp/log4go"
)

// InitSignal register signals handler.
func InitSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP)
	for {
		s := <-c
		log.Info("comet[%s] get a signal %s", Ver, s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGSTOP, syscall.SIGINT:
			return
		case syscall.SIGHUP:
			reload()
		case syscall.SIGTERM:
			shutdown()
		default:
			return
		}
	}
}

func reload() {
	newConf, err := ReloadConfig()
	if err != nil {
		log.Error("ReloadConfig() error(%v)", err)
		return
	}
	Conf = newConf
	Debug = Conf.Debug
	updateServerID()
	updateWhitelist()
}

func updateWhitelist() {
	wl, err := NewWhitelist(Conf.WhiteLog, Conf.Whitelist)
	if err != nil {
		log.Error("NewWhitelist() error(%v)", err)
		return
	}
	DefaultWhitelist = wl
}

func shutdown() {
	DefaultServer.InShutdown = true
	ShutdownTCP()
	ShutdownWebsocket()
	log.Warn("Server is shutting down... no new connections will be accepted")
}