package main

import (
	"log"

	"github.com/Gealber/adnl-tcp/client"
	"github.com/Gealber/adnl-tcp/config"
)

func main() {
	cfg := config.Config()
	clt, err := client.New(cfg)
	if err != nil {
		panic(err)
	}

	filename := "global.config.json"
	globalNetConfig, err := clt.LoadGlobalConfig(filename)
	if err != nil {
		panic(err)
	}

	if len(globalNetConfig.LiteServers) == 0 {
		panic("no liteservers on global config")
	}

    lts := globalNetConfig.LiteServers[0]

	log.Printf("Trying to connect to liteserver: %s....\n", lts.ConnStr())
	// let's test this with our local setup
	err = clt.Connect(lts)
	if err != nil {
		panic(err)
	}

	if err := clt.Close(); err != nil {
		panic(err)
	}
}
