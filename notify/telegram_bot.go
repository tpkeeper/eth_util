package notify

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/tpkeeper/eth-util/db"
	"log"
	"strconv"
)

const (
	addMonitorStep              = "addMonitor"
	addContractAddressStep      = "addContractAddress"
	deleteMonitorStep           = "deleteMonitor"
	deleteByContractAddressStep = "deleteContractAddress"
	deleteByTokenAddressStep    = "deleteTokenAddress"
	listMonitorStep             = "listMonitor"
)

type Stack []string

func (s *Stack) Pop() (string, error) {
	if len(*s) == 0 {
		return "", fmt.Errorf("empty")
	}
	r := (*s)[len(*s)-1]
	*s = (*s)[0 : len(*s)-1]
	return r, nil
}

func (s *Stack) Push(e string) {
	*s = append(*s, e)
}

func (s *Stack) Top() string {
	if len(*s) == 0 {
		return ""
	}
	return (*s)[len(*s)-1]
}

func (s *Stack) Clear() {
	*s = (*s)[:0]
}

var (
	step                = new(Stack) //store pre step
	tempContractAddress string
	mainMenu            = tgbotapi.NewInlineKeyboardMarkup(
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
			tgbotapi.NewInlineKeyboardButtonData("Delete by contract address", deleteByContractAddressStep),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Delete by token address", deleteByTokenAddressStep),
		),
	)
)

type TelegramBot struct {
	*tgbotapi.BotAPI
	db *db.Db
}

func NewTelegramBot(token string, db *db.Db) *TelegramBot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)
	return &TelegramBot{bot, db}
}

func (tg *TelegramBot) Notify(chatId string, msg string) error {
	id, err := strconv.ParseInt(chatId, 10, 64)
	if err != nil {
		return err
	}
	retMsg := tgbotapi.NewMessage(id, msg)
	_, err = tg.Send(retMsg)

	return err
}

func (tg *TelegramBot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := tg.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			if update.Message.IsCommand() {
				tg.handleMessageCommand(update.Message)
			} else {
				tg.handleMessageText(update.Message)
			}
		}
		if update.CallbackQuery != nil {
			tg.handleCallbackQuery(update.CallbackQuery)
		}
	}
}

func (tg *TelegramBot) handleMessageCommand(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	switch message.Command() {
	case "start":
		msg.Text = "hello! I`m tpkeeper`s bot.this is the main menu:\n"
		msg.ReplyMarkup = mainMenu
	default:
		msg.Text = "sorry,this command not exist!"
	}
	tg.Send(msg)
}

func (tg *TelegramBot) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	tg.AnswerCallbackQuery(tgbotapi.NewCallback(callbackQuery.ID, callbackQuery.Data))
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "")
	switch callbackQuery.Data {
	case addMonitorStep:
		msg.Text = "please enter contract address:"
		_, err := tg.Send(msg)
		if err == nil {
			step.Clear()
			step.Push(addMonitorStep)
		}
	case deleteMonitorStep:
		msg.Text = "please select one:"
		msg.ReplyMarkup = deleteMenu
		_, err := tg.Send(msg)
		if err == nil {
			step.Clear()
			step.Push(deleteMonitorStep)
		}
	case deleteByContractAddressStep:
		if step.Top() != deleteMonitorStep {
			msg.Text = "bad menu, return main menu:"
			msg.ReplyMarkup = mainMenu
			tg.Send(msg)
			step.Clear()
			return
		}

		msg.Text = "please enter contract address:"
		_, err := tg.Send(msg)
		if err == nil {
			step.Push(deleteByContractAddressStep)
		}
	case deleteByTokenAddressStep:
		if step.Top() != deleteMonitorStep {
			msg.Text = "bad menu, return main menu:"
			msg.ReplyMarkup = mainMenu
			tg.Send(msg)
			step.Clear()
			return
		}
		msg.Text = "please enter token address:"
		_, err := tg.Send(msg)
		if err == nil {
			step.Push(deleteByTokenAddressStep)
		}
	case listMonitorStep:
	default:
		step.Clear()
		msg.Text = "bad menu, return main menu:"
		msg.ReplyMarkup = mainMenu
		tg.Send(msg)
	}
}

func (tg *TelegramBot) handleMessageText(message *tgbotapi.Message) {
	retMsg := tgbotapi.NewMessage(message.Chat.ID, "")

	switch step.Top() {
	case addMonitorStep:
		tempContractAddress = message.Text
		retMsg.Text = "please enter token address:"
		_, err := tg.Send(retMsg)
		if err == nil {
			step.Push(addContractAddressStep)
		}
	case addContractAddressStep:
		monitorTargetErc20s, err := tg.db.GetMonitorTargetErc20sFromDb()
		if err != nil {
			fmt.Println(err)
			step.Clear()
			retMsg.Text = "bad step, return main menu:"
			retMsg.ReplyMarkup = mainMenu
			tg.Send(retMsg)
			return
		}
		key := tempContractAddress + message.Text
		if monitorTarget, exist := monitorTargetErc20s[key]; exist {
			monitorTarget.ChatId[strconv.FormatInt(message.Chat.ID, 10)] = struct{}{}
		} else {
			chatIdMap := make(map[string]struct{})
			chatIdMap[strconv.FormatInt(message.Chat.ID, 10)] = struct{}{}
			monitorTargetErc20s[key] = &db.MonitorTargetErc20{
				ContractAddress: tempContractAddress, TokenAddress: message.Text, ChatId: chatIdMap}
		}

		err = tg.db.SaveMonitorTargetErc20ToDb(*monitorTargetErc20s[key])
		if err != nil {
			fmt.Printf("saveMonitorTargetToDb %s", err)
		}

		retMsg.Text = fmt.Sprintf("add monitor ok! \ncontract address: %s \ntoken address: %s\n",
			tempContractAddress, message.Text)
		_, err = tg.Send(retMsg)
		if err == nil {
			step.Clear()
		}
	case deleteByContractAddressStep:
		monitorTargetErc20s, err := tg.db.GetMonitorTargetErc20sFromDb()
		if err != nil {
			fmt.Println(err)
			step.Clear()
			retMsg.Text = "bad step, return main menu:"
			retMsg.ReplyMarkup = mainMenu
			tg.Send(retMsg)
			return
		}
		contractAddr := message.Text
		chatIdStr := strconv.FormatInt(message.Chat.ID, 10)
		for key, monitorTargetErc20 := range monitorTargetErc20s {
			if monitorTargetErc20.ContractAddress == contractAddr {
				if _, exist := monitorTargetErc20.ChatId[chatIdStr]; exist {
					delete(monitorTargetErc20.ChatId, chatIdStr)
					err := tg.db.SaveMonitorTargetErc20ToDb(*monitorTargetErc20)
					if err != nil {
						fmt.Println(err)
					}
				}
				if len(monitorTargetErc20.ChatId) == 0 {
					delete(monitorTargetErc20s, key)
					err := tg.db.DelMonitorTargetErc20FromDb(key)
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		retMsg.Text = fmt.Sprintf("delete monitor ok! \ncontract address: %s\n", contractAddr)
		_, err = tg.Send(retMsg)
		if err == nil {
			step.Clear()
		}

	case deleteByTokenAddressStep:
		monitorTargetErc20s, err := tg.db.GetMonitorTargetErc20sFromDb()
		if err != nil {
			fmt.Println(err)
			step.Clear()
			retMsg.Text = "bad step, return main menu:"
			retMsg.ReplyMarkup = mainMenu
			tg.Send(retMsg)
			return
		}
		tokenAddr := message.Text
		chatIdStr := strconv.FormatInt(message.Chat.ID, 10)

		for key, monitorTargetErc20 := range monitorTargetErc20s {
			if monitorTargetErc20.TokenAddress == tokenAddr {
				if _, exist := monitorTargetErc20.ChatId[chatIdStr]; exist {
					delete(monitorTargetErc20.ChatId, chatIdStr)
					err := tg.db.SaveMonitorTargetErc20ToDb(*monitorTargetErc20)
					if err != nil {
						fmt.Println(err)
					}
				}
				if len(monitorTargetErc20.ChatId) == 0 {
					delete(monitorTargetErc20s, key)
					err := tg.db.DelMonitorTargetErc20FromDb(key)
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		retMsg.Text = fmt.Sprintf("delete monitor ok! \ntoken address: %s\n", tokenAddr)
		_, err = tg.Send(retMsg)
		if err == nil {
			step.Clear()
		}

	case listMonitorStep:
	default:
		step.Clear()
		retMsg.Text = "bad step, return main menu:"
		retMsg.ReplyMarkup = mainMenu
		tg.Send(retMsg)
	}
}
