package main

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"math/big"
)

const (
	addMonitorStep            = "addMonitor"
	addContractAddressStep    = "addContractAddress"
	addTokenAddressStep       = "addTokenAddress"
	deleteMonitorStep         = "deleteMonitor"
	deleteContractAddressStep = "deleteContractAddress"
	deleteTokenAddressStep    = "deleteTokenAddress"
	listMonitorStep           = "listMonitor"
)

var (
	monitorTargetErc20s map[string]*MonitorTargetErc20 // ContractAddress+TokenAddress -> []MonitorTargetErc20
)

type MonitorTargetErc20 struct {
	ContractAddress string
	TokenAddress    string
	Amount          big.Int             `json:"-"`
	ChatId          map[string]struct{} //key is string so it is easy to marshal
}

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

	db, err := newDb(dbFilePath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	monitorTargetErc20s, err = db.GetMonitorTargetErc20sFromDb()
	if err != nil {
		panic(err)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.TgToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	go monitor(ctx, bot)

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			if update.Message.IsCommand() {
				handleMessageCommand(bot, update.Message)
			} else {
				if len(step) == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
					msg.Text = "not a command"
					bot.Send(msg)
					continue
				}
				handleMessageText(bot, db, update.Message)
			}
		}
		if update.CallbackQuery != nil {
			handleCallbackQuery(bot, update.CallbackQuery)
		}
	}
}
