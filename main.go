package main

import (
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tomasen/trafcacc/src"

	"net/http"
	_ "net/http/pprof"

	log "github.com/Sirupsen/logrus"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// -listen=tcp://:500 -upstream=udp://172.0.0.1:2000-2100
	// -listen=udp://:2000-2100 -upstream=tcp://172.0.0.1:500
	listen := flag.String("listen", "<proto>://<ip>:<port begin-end>[,...] eg. udp://0.0.0.0:500", "listen to")
	upstream := flag.String("upstream", "<proto>://<ip>:<port begin-end>[,...] eg. udp://172.0.0.1:2000-2100,192.168.1.1:2000-2050", "send to")
	role := flag.String("role", "frontend", "work as backend or frontend")
	loglevel := flag.Bool("v", false, "set log level to debug")
	pprof := flag.String("pprof", "", "pprof listen to")

	flag.Parse()

	if *loglevel {
		log.SetLevel(log.DebugLevel)
	}

	var t trafcacc.Trafcacc
	switch *role {
	case "backend":
		t = trafcacc.Accelerate(*listen, *upstream, trafcacc.BACKEND)
	default:
		t = trafcacc.Accelerate(*listen, *upstream, trafcacc.FRONTEND)
	}

	if len(*pprof) != 0 {
		go func() {
			log.Println(http.ListenAndServe(*pprof, nil))
		}()
	}

	go func() {
		ct := time.Tick(3 * time.Second)
		for _ = range ct {
			t.PrintStatus()
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	<-c
	// cleanup
	os.Exit(1)
}
