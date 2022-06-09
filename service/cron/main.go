package main

import (
	"context"
	"dcard-pretest/pkg/store"
	"log"
	"os"
	"os/signal"

	"github.com/robfig/cron/v3"
	logger "github.com/sirupsen/logrus"
)

const (
	leaderboardKey = "dcard-leaderboard"
)

func main() {
	log.Println("Starting cron ....")

	c := cron.New(cron.WithSeconds())

	c.AddFunc("@every 600s", resetLeaderboard)

	c.Start()
	defer c.Stop()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig
}

func resetLeaderboard() {
	ctx := context.TODO()

	err := store.NewRedis().Reset(ctx, leaderboardKey)
	if err != nil {
		logger.Errorln("fail to reset leaderboard", err)
	}
	logger.Infoln("success to reset leaderboard", leaderboardKey)
}
