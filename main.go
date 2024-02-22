package main

import (
	"context"
	"encoding/hex"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/astriaorg/messenger-rollup/messenger"
	"github.com/sethvargo/go-envconfig"
)

type HexFormatter struct {
	log.TextFormatter
}

func (f *HexFormatter) Format(entry *log.Entry) ([]byte, error) {
	// Convert any byte slice fields to hex strings
	for key, value := range entry.Data {
		if data, ok := value.([]byte); ok {
			entry.Data[key] = fmt.Sprintf("0x%s", hex.EncodeToString(data))
		}
	}

	// Use the embedded TextFormatter to continue normal formatting
	return f.TextFormatter.Format(entry)
}

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&HexFormatter{log.TextFormatter{}})

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
