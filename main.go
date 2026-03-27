package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"quiz-race/engine"
	"quiz-race/mock"
	"quiz-race/server"
)

const (
	defaultUsers = 1000
	serverAddr   = ":8080"
	bootDelay    = 100 * time.Millisecond
)

func main() {
	n := flag.Int("n", defaultUsers, "number of simulated users")
	flag.Parse()

	e := engine.NewGameEngine()
	s := server.NewServer(e)

	go func() {
		if err := s.Start(serverAddr); err != nil {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		}
	}()

	time.Sleep(bootDelay)
	mock.Run(*n, "http://localhost"+serverAddr)
}
