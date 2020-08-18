package main

import (
	"context"
	"github.com/BurntSushi/toml"
	"github.com/tpkeeper/eth-util/db"
	"github.com/tpkeeper/eth-util/log"
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
	log.Logger.Info().Str("tgToken",cfg.TgToken).Str("etherscanKey",cfg.EtherscanKey).Msg("config")

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
