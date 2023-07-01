package main

import (
	"Ethereum_Service/config"
	"Ethereum_Service/internal/services/indexer_service"
	"flag"
	"os"
	"os/signal"
	"syscall"
)

var (
	// flagconf is the config flag.
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../", "config path, eg: -conf config.yaml")
}

func handleSignals(service *indexer_service.Service) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	service.Shutdown()
}

func main() {
	flag.Parse()
	config.LoadConf(flagconf, config.GetConfig())

	service, err := indexer_service.NewService(config.GetConfig().RCPEndpoint)

	if err != nil {
		panic(err)
	}
	go service.Start(config.GetConfig().WorkerNumber)

	handleSignals(service)
}
