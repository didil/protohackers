package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/didil/protohackers/server"
	"github.com/didil/protohackers/services"
	"go.uber.org/zap"
)

func main() {
	mode := flag.String("m", "", "protohackers mode")
	port := flag.Int("p", 3000, "port to listen to")
	flag.Parse()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("logger init failed %v", err)
	}
	defer logger.Sync() // flushes buffer, if any

	chatSvc := services.NewChatService(logger)
	unusualDbSvc := services.NewUnusualDbService()

	s, err := server.NewServer(*mode, *port, logger, server.WithChatService(chatSvc), server.WithUnusualDbService(unusualDbSvc))
	if err != nil {
		logger.Fatal("server init failed", zap.Error(err))
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan bool, 1)

	go func() {
		sig := <-sigs
		logger.Sugar().Infof("Received signal: %v. Exiting ...", sig)
		close(done)
	}()

	err = s.Start(done)
	if err != nil {
		logger.Fatal("server start error", zap.Error(err))
	}

	os.Exit(0)
}
