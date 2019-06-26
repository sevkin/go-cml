package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	const envURL = "CML_URL"

	// flag.Parse()

	// log.SetFlags(0)

	pinger, err := newPinger(os.Getenv(envURL))
	if err != nil {
		log.Fatalf("can`t init pinger with err: %v", err)
	}

	term := make(chan struct{}, 1)

	go func() {
		sigINT := make(chan os.Signal, 2)
		signal.Notify(sigINT, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-sigINT:
			log.Printf("terminated by signal: %v", sig.String())
			close(term)
		}
	}()

	pinger.ping(term)

	log.Print("all done, bye bye")
}
