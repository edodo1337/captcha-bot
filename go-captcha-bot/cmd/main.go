package main

import (
	"captcha-bot/internal/app/utils"
	"captcha-bot/internal/pkg/conf"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config := conf.New()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Println("Starting Bot")
		if err := utils.RunPolling(ctx, config); err != nil {
			log.Fatal(err.Error())

			return
		}
	}()

	<-stop

	log.Println("Bot shutdown")
}
