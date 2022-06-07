package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/robfig/cron/v3"
)

type leaderboard struct {
}

func (r leaderboard) Reset() error {
	log.Println("clear leaderboard...")
	return nil
}
func resetLeaderboard() {
	l := &leaderboard{}
	l.Reset()
}

func main() {
	log.Println("Starting cron ....")

	c := cron.New(cron.WithSeconds())

	// TODO: modify to every 10 min
	c.AddFunc("@every 5s", resetLeaderboard)

	c.Start()
	defer c.Stop()

	// TODO: graceful shutdown
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig
}
