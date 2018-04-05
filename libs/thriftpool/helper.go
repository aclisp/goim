package thriftpool

import (
	"time"

	log "github.com/thinkboy/log4go"
)

// Run accept the generic service client
type Run func(interface{}) error

func Invoke(prefix string, pool Pool, run Run, factory Factory) (err error) {
	conn, err := pool.Get()
	if err != nil {
		log.Error("%q can not get thrift conn from pool: %v", prefix, err)
		return
	}
	defer conn.Close()
	err = run(conn.Client)
	if err != nil {
		conn.Conn.Close()                           // close the socket that failed
		if conn.Conn, err = factory(); err != nil { // reconnect the socket
			log.Error("%q reconnect failed, server down? %v", prefix, err)
			conn.MarkUnusable()
			return
		}
		err = run(conn.Client) // retry on the newly connected socket
		if err != nil {
			log.Error("%q failed after reconnect, fatal! %v", prefix, err)
			conn.MarkUnusable()
			return
		}
	}
	return
}

func Ping(prefix string, pool Pool, run Run, interval time.Duration) {
	var count int
	for {
		count = 0
		n := pool.Len()
		for i := 0; i < n; i++ {
			conn, err := pool.Get()
			if err != nil {
				break
			}
			err = run(conn.Client)
			if err != nil {
				count++
				conn.MarkUnusable()
			}
			conn.Close()
		}
		if count > 0 {
			log.Info("%q removed %d stale thrift connection(s) out of %d in the pool", prefix, count, n)
		}
		time.Sleep(interval)
	}
}
