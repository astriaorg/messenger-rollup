package main

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/astriaorg/messenger-rollup/messenger"
	"github.com/sethvargo/go-envconfig"
)

func main() {
	log.SetLevel(log.DebugLevel)

	// load env vars
	var cfg messenger.Config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		log.Fatal(err)
	}
	log.Debugf("Read config from env: %+v\n", cfg)

	// init from cfg
	app := messenger.NewApp(cfg)

	// run messenger
	app.Run()
}
