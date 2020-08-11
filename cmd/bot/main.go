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
	step                []string
	monitorTargetErc20s = make(map[string]*MonitorTargetErc20) // contractAddress+tokenAddress -> []MonitorTargetErc20

	tempContractAddress string

	mainMenu = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Add Monitor", addMonitorStep),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Delete Monitor", deleteMonitorStep),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("List Monitor", listMonitorStep),
		),
	)

	deleteMenu = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Delete by contract address", deleteContractAddressStep),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Delete by token address", deleteTokenAddressStep),
		),
	)
)

type MonitorTargetErc20 struct {
	contractAddress string
	tokenAddress    string
	amount          big.Int
	chatId          map[string]struct{} //key is string so it is easy to marshal
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
				handleMessageText(bot, update.Message)
			}
		}

		if update.CallbackQuery != nil {
			handleCallbackQuery(bot, update.CallbackQuery)
		}
	}
}
