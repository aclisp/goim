package main

import (
	"flag"
	"goim/libs/io/ioutil"
	"runtime"

	log "github.com/aclisp/log4go"
)

func main() {
	flag.Parse()
	if err := InitConfig(); err != nil {
		panic(err)
	}
	log.LoadConfiguration(Conf.Log)
	runtime.GOMAXPROCS(runtime.NumCPU())
	//comet
	err := InitComet(Conf.Comets,
		CometOptions{
			RoutineSize: Conf.RoutineSize,
			RoutineChan: Conf.RoutineChan,
		})
	if err != nil {
		log.Warn("comet rpc current can't connect, retry")
	}
	// start monitor
	if Conf.MonitorOpen {
		InitMonitor(Conf.MonitorAddrs)
	}
	//round
	round := NewRound(RoundOptions{
		Timer:     Conf.Timer,
		TimerSize: Conf.TimerSize,
	})
	//room
	InitRoomBucket(round,
		RoomOptions{
			BatchNum:   Conf.RoomBatch,
			SignalTime: Conf.RoomSignal,
		})
	//room info
	MergeRoomServers()
	go SyncRoomServers()
	InitPush()
	if Conf.KafkaOpen {
		if err := InitKafka(); err != nil {
			panic(err)
		}
	}
	if err := InitRPC(); err != nil {
		panic(err)
	}
	if err := ioutil.WritePidFile(Conf.PidFile); err != nil {
		panic(err)
	}
	// block until a signal is received.
	InitSignal()
}
