package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
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
	contractTarget      = make(map[string][]string)
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
	tgToken      string
	etherscanKey string
}

func main() {
	cfg := config{}
	_, err := toml.DecodeFile("./conf.toml", &cfg)
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg)
	bot, err := tgbotapi.NewBotAPI(cfg.tgToken)
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
		contractTarget[tempContractAddress] = append(contractTarget[tempContractAddress],
			message.Text)

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
