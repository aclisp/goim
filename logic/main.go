package main

import (
	"flag"
	"fmt"
	"goim/libs/io/ioutil"
	"goim/libs/perf"
	"runtime"

	log "github.com/aclisp/log4go"
)

func main() {
	flag.Parse()
	if err := InitConfig(); err != nil {
		panic(err)
	}
	runtime.GOMAXPROCS(Conf.MaxProc)
	log.LoadConfiguration(Conf.Log)
	defer log.Close()
	log.Info("logic[%s] start", Ver)
	perf.Init(Conf.PprofAddrs)
	// router rpc
	if err := InitRouter(Conf.RouterRPCAddrs); err != nil {
		log.Warn("router rpc current can't connect, retry")
	}
	// start monitor
	if Conf.MonitorOpen {
		InitMonitor(Conf.MonitorAddrs)
	}
	MergeCount()
	go SyncCount()
	// logic rpc
	autherBuilder, ok := AutherRegistry[Conf.AuthMode]
	if !ok {
		panic(fmt.Errorf("unknown auth mode: %s", Conf.AuthMode))
	} else {
		log.Info("auth mode: %s", Conf.AuthMode)
	}
	if err := InitRPC(autherBuilder()); err != nil {
		panic(err)
	}
	// to job
	if err := InitJobRpc(Conf.JobAddrs); err != nil {
		log.Warn("job rpc current can't connect, retry")
	}
	if err := InitHTTP(); err != nil {
		panic(err)
	}
	if Conf.KafkaOpen {
		if err := InitKafka(Conf.KafkaAddrs); err != nil {
			panic(err)
		}
	}
	if err := ioutil.WritePidFile(Conf.PidFile); err != nil {
		panic(err)
	}
	// block until a signal is received.
	InitSignal()
}
