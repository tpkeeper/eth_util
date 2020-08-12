package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
)

const dbFilePath = "./bot.db"

func handleMessageText(bot *tgbotapi.BotAPI, db *Db, message *tgbotapi.Message) {
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
			monitorTarget.ChatId[strconv.FormatInt(message.Chat.ID, 10)] = struct{}{}
		} else {
			chatIdMap := make(map[string]struct{})
			chatIdMap[strconv.FormatInt(message.Chat.ID, 10)] = struct{}{}
			monitorTargetErc20s[key] = &MonitorTargetErc20{
				ContractAddress: tempContractAddress, TokenAddress: message.Text, ChatId: chatIdMap}
		}

		err := db.SaveMonitorTargetErc20ToDb(*monitorTargetErc20s[key])
		if err != nil {
			fmt.Printf("saveMonitorTargetToDb %s", err)
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

func handleMessageCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
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

func handleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
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
