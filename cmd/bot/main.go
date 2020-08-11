package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"time"
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

type MonitorTargetErc20 struct {
	contractAddress string
	tokenAddress    string
	amount          *big.Int
	chatId          map[int64]struct{}
}

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
				dealMessageCommand(bot, update.Message)
			} else {
				if len(step) == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
					msg.Text = "not a command"
					bot.Send(msg)
					continue
				}
				dealMessage(bot, update.Message)
			}
		}

		if update.CallbackQuery != nil {
			dealCallbackQuery(bot, update.CallbackQuery)
		}
	}
}

type TokenBalanceRes struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

func monitor(ctx context.Context, bot *tgbotapi.BotAPI) {
	ticker := time.NewTicker(time.Second * 5)
	api := "https://api.etherscan.io/api?module=account&action=tokenbalance&contractaddress=%s&address=%s&tag=latest&apikey=%s"
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			for _, tokenInfo := range monitorTargetErc20s {
				res, err := http.Get(fmt.Sprintf(api, tokenInfo.contractAddress, tokenInfo.tokenAddress, "VC35I1VEW49ZTRNPDT11QQ8WWCS324FZGS"))
				if err != nil {
					fmt.Println(err)
					continue
				}
				tokenBalanceRes := TokenBalanceRes{}
				bts, err := ioutil.ReadAll(res.Body)
				if err != nil {
					fmt.Println(err)
					continue
				}
				err = json.Unmarshal(bts, &tokenBalanceRes)
				if err != nil {
					fmt.Println(err)
					continue
				}
				if tokenBalanceRes.Status != "1" {
					fmt.Println(tokenBalanceRes)
					continue
				}
				balanceBitInt := new(big.Int)
				balanceBitInt.SetString(tokenBalanceRes.Result, 10)
				balanceBitInt.Div(balanceBitInt, big.NewInt(1000000000000000000))

				for chatId, _ := range tokenInfo.chatId {
					msg := tgbotapi.NewMessage(chatId, tokenBalanceRes.Result)
					_, err := bot.Send(msg)
					if err != nil {
						fmt.Println(err)
					}
				}
				tokenInfo.amount = balanceBitInt

			}
		case <-ctx.Done():
			return
		}
	}
}


func dealMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	switch step[len(step)-1] {
	case addContractAddressStep:

		tempContractAddress = message.Text
		msg.Text = "please enter token address"
		_, err := bot.Send(msg)
		if err == nil {
			step = append(step, addTokenAddressStep)
		}

	case addTokenAddressStep:

		step = append(step, addTokenAddressStep)
		key := tempContractAddress + message.Text
		if monitorTarget, exist := monitorTargetErc20s[key]; exist {
			monitorTarget.chatId[message.Chat.ID] = struct{}{}
		} else {
			chatIdMap := make(map[int64]struct{})
			chatIdMap[message.Chat.ID] = struct{}{}
			monitorTargetErc20s[key] = &MonitorTargetErc20{
				contractAddress: tempContractAddress, tokenAddress: message.Text, chatId: chatIdMap}
		}

		msg.Text = fmt.Sprintf("add monitor ok!\ncontract address: %s \ntoken address: %s",
			tempContractAddress, message.Text)
		bot.Send(msg)
	case deleteMonitorStep:
	case listMonitorStep:
	default:
		step = []string{}
		msg.Text = "bad step, return main menu"
		msg.ReplyMarkup = mainMenu
		bot.Send(msg)
	}
}

func dealMessageCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	switch message.Command() {
	case "start":
		msg.Text = "hello! I`m tpkeeper`s bot.\n/menu to show the main menu"
	case "menu":
		msg.Text = "main menu:"
		msg.ReplyMarkup = mainMenu
	default:
		msg.Text = "sorry,this command not exist!"
	}
	bot.Send(msg)
}

func dealCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	bot.AnswerCallbackQuery(tgbotapi.NewCallback(callbackQuery.ID, callbackQuery.Data))
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "")
	switch callbackQuery.Data {
	case addMonitorStep:
		msg.Text = "please enter contract address"
		_, err := bot.Send(msg)
		if err == nil {
			step = append(step, addContractAddressStep)
		}
	case deleteMonitorStep:
		msg.Text = "please select one:"
		msg.ReplyMarkup = deleteMenu
		_, err := bot.Send(msg)
		if err != nil {
		}
	case listMonitorStep:
	default:
		step = []string{}
		msg.Text = "bad step, return main menu"
		msg.ReplyMarkup = mainMenu
		bot.Send(msg)
	}
}