package main

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/tpkeeper/eth-util/db"
	"github.com/tpkeeper/eth-util/monitor"
	"github.com/tpkeeper/eth-util/notify"
)

const dbFilePath = "./bot.db"


type config struct {
	TgToken      string
	EtherscanKey string
}

func main() {
	cfg := config{}
	_, err := toml.DecodeFile("./conf.toml", &cfg)
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg)

	db, err := db.NewDb(dbFilePath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	tgBot := notify.NewTelegramBot(cfg.TgToken, db)
	erc20Monitor := monitor.NewErc20Monitor(db, []notify.Notifier{tgBot})

	go erc20Monitor.Start(ctx)
	go tgBot.Start()

	select {}
}
