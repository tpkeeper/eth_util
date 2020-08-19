package notify

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/tpkeeper/eth-util/db"
	"github.com/tpkeeper/eth-util/log"
	"strconv"
	"strings"
)

const (
	addMonitorStep              = "addMonitor"
	addContractAddressStep      = "addContractAddress"
	deleteMonitorStep           = "deleteMonitor"
	deleteByContractAddressStep = "deleteContractAddress"
	deleteByTokenAddressStep    = "deleteTokenAddress"
	listMonitorStep             = "listMonitor"

	warnBedStep = "sorry,you are in bad step! now return to main menu:"
)

var (
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
		log.Logger.Err(err).
			Str("token", token).
			Send()
	}
	bot.Debug = false
	log.Logger.Info().Msg(fmt.Sprintf("Authorized on account %s", bot.Self.UserName))
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
		msg.Text = "sorry,this command doesn`t  exist!\n" +
			"the command list :\n" +
			"1: /start"
	}
	tg.Send(msg)
}

func (tg *TelegramBot) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	chatIdStr := strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)
	step, err := tg.db.GetStepFromDb(chatIdStr)
	if err != nil {
		log.Logger.Err(err).Send()
		return
	}
	tg.AnswerCallbackQuery(tgbotapi.NewCallback(callbackQuery.ID, callbackQuery.Data))
	retMsg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "")
	switch callbackQuery.Data {
	case addMonitorStep:
		retMsg.Text = "please enter contract address:"
		_, err := tg.Send(retMsg)
		if err == nil {
			step.Clear()
			step.Push(addMonitorStep)
		}
	case deleteMonitorStep:
		retMsg.Text = "please select one:"
		retMsg.ReplyMarkup = deleteMenu
		_, err := tg.Send(retMsg)
		if err == nil {
			step.Clear()
			step.Push(deleteMonitorStep)
		}
	case deleteByContractAddressStep:
		if step.Top() != deleteMonitorStep {
			retMsg.Text = warnBedStep
			retMsg.ReplyMarkup = mainMenu
			tg.Send(retMsg)
			step.Clear()
		} else {
			retMsg.Text = "please enter contract address:"
			_, err := tg.Send(retMsg)
			if err == nil {
				step.Push(deleteByContractAddressStep)
			}
		}
	case deleteByTokenAddressStep:
		if step.Top() != deleteMonitorStep {
			retMsg.Text = warnBedStep
			retMsg.ReplyMarkup = mainMenu
			tg.Send(retMsg)
			step.Clear()
		} else {
			retMsg.Text = "please enter token address:"
			_, err := tg.Send(retMsg)
			if err == nil {
				step.Push(deleteByTokenAddressStep)
			}
		}
	case listMonitorStep:
		monitorTargetErc20s, err := tg.db.GetMonitorTargetErc20sFromDb()
		if err != nil {
			log.Logger.Err(err).Send()
			step.Clear()
			retMsg.Text = "get list err, return main menu:"
			retMsg.ReplyMarkup = mainMenu
			tg.Send(retMsg)
		} else {
			chatIdStr := strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)
			var strBuilder strings.Builder
			tmp := make(map[string][]string)

			for _, monitorTargetErc20 := range monitorTargetErc20s {
				if _, exist := monitorTargetErc20.ChatId[chatIdStr]; exist {
					tmp[monitorTargetErc20.ContractAddress] = append(tmp[monitorTargetErc20.ContractAddress],
						monitorTargetErc20.TokenAddress)
				}
			}

			for contractAddr, tokenAddrs := range tmp {
				strBuilder.WriteString(fmt.Sprintf("\ncontract: %s\n", contractAddr))
				for _, addr := range tokenAddrs {
					strBuilder.WriteString(fmt.Sprintf("->token address: %s\n", addr))
				}
			}

			retMsg.Text = strBuilder.String()
			_, err = tg.Send(retMsg)
			if err == nil {
				step.Clear()
			}
		}
	default:
		step.Clear()
		retMsg.Text = warnBedStep
		retMsg.ReplyMarkup = mainMenu
		tg.Send(retMsg)
	}

	err = tg.db.SaveStepToDb(chatIdStr, step)
	if err != nil {
		log.Logger.Err(err).Send()
	}
}

func (tg *TelegramBot) handleMessageText(message *tgbotapi.Message) {
	chatIdStr := strconv.FormatInt(message.Chat.ID, 10)
	step, err := tg.db.GetStepFromDb(chatIdStr)
	if err != nil {
		log.Logger.Err(err).Send()
		return
	}
	retMsg := tgbotapi.NewMessage(message.Chat.ID, "")

	//validate address when in step
	if len(step) != 0 && !IsHexAddress(message.Text) {
		step.Clear()
		retMsg.Text = "not a hex address! now return to main menu:"
		retMsg.ReplyMarkup = mainMenu
		tg.Send(retMsg)
		//save step before return
		err = tg.db.SaveStepToDb(chatIdStr, step)
		if err != nil {
			log.Logger.Err(err).Send()
		}
		return
	}
	//save lower case
	message.Text = strings.ToLower(message.Text)

	switch step.Top() {
	case addMonitorStep:
		err := tg.db.SaveTempContractAddrToDb(chatIdStr, message.Text)
		if err != nil {
			log.Logger.Err(err).Send()
			step.Clear()
			retMsg.Text = warnBedStep
			retMsg.ReplyMarkup = mainMenu
			tg.Send(retMsg)
		} else {
			retMsg.Text = "please enter token address:"
			_, err = tg.Send(retMsg)
			if err == nil {
				step.Push(addContractAddressStep)
			}
		}
	case addContractAddressStep:
		monitorTargetErc20s, err := tg.db.GetMonitorTargetErc20sFromDb()
		tempContractAddress, err := tg.db.GetTempContractAddrFromDb(chatIdStr)
		if err != nil {
			log.Logger.Err(err).Send()
			step.Clear()
			retMsg.Text = warnBedStep
			retMsg.ReplyMarkup = mainMenu
			tg.Send(retMsg)
		} else {

			key := tempContractAddress + message.Text
			if monitorTarget, exist := monitorTargetErc20s[key]; exist {
				monitorTarget.ChatId[strconv.FormatInt(message.Chat.ID, 10)] = struct{}{}
			} else {
				chatIdMap := make(map[string]struct{})
				chatIdMap[chatIdStr] = struct{}{}
				monitorTargetErc20s[key] = &db.MonitorTargetErc20{
					ContractAddress: tempContractAddress, TokenAddress: message.Text, ChatId: chatIdMap}
			}

			err = tg.db.SaveMonitorTargetErc20ToDb(*monitorTargetErc20s[key])
			if err != nil {
				log.Logger.Err(err).Send()
			}

			retMsg.Text = fmt.Sprintf("add monitor ok! \ncontract address: %s \ntoken address: %s\n",
				tempContractAddress, message.Text)
			_, err = tg.Send(retMsg)
			if err == nil {
				step.Clear()
			}
		}
	case deleteByContractAddressStep:
		monitorTargetErc20s, err := tg.db.GetMonitorTargetErc20sFromDb()
		if err != nil {
			log.Logger.Err(err).Send()
			step.Clear()
			retMsg.Text = warnBedStep
			retMsg.ReplyMarkup = mainMenu
			tg.Send(retMsg)
		} else {
			contractAddr := message.Text
			for key, monitorTargetErc20 := range monitorTargetErc20s {
				if monitorTargetErc20.ContractAddress == contractAddr {
					if _, exist := monitorTargetErc20.ChatId[chatIdStr]; exist {
						delete(monitorTargetErc20.ChatId, chatIdStr)
						err := tg.db.SaveMonitorTargetErc20ToDb(*monitorTargetErc20)
						if err != nil {
							log.Logger.Err(err).Send()
						}
					}
					if len(monitorTargetErc20.ChatId) == 0 {
						delete(monitorTargetErc20s, key)
						err := tg.db.DelMonitorTargetErc20FromDb(key)
						if err != nil {
							log.Logger.Err(err).Send()
						}
					}
				}
			}

			retMsg.Text = fmt.Sprintf("delete monitor ok! \ncontract address: %s\n", contractAddr)
			_, err = tg.Send(retMsg)
			if err == nil {
				step.Clear()
			}
		}

	case deleteByTokenAddressStep:
		monitorTargetErc20s, err := tg.db.GetMonitorTargetErc20sFromDb()
		if err != nil {
			log.Logger.Err(err).Send()
			step.Clear()
			retMsg.Text = warnBedStep
			retMsg.ReplyMarkup = mainMenu
			tg.Send(retMsg)
		} else {
			tokenAddr := message.Text
			for key, monitorTargetErc20 := range monitorTargetErc20s {
				if monitorTargetErc20.TokenAddress == tokenAddr {
					if _, exist := monitorTargetErc20.ChatId[chatIdStr]; exist {
						delete(monitorTargetErc20.ChatId, chatIdStr)
						err := tg.db.SaveMonitorTargetErc20ToDb(*monitorTargetErc20)
						if err != nil {
							log.Logger.Err(err).Send()
						}
					}
					if len(monitorTargetErc20.ChatId) == 0 {
						delete(monitorTargetErc20s, key)
						err := tg.db.DelMonitorTargetErc20FromDb(key)
						if err != nil {
							log.Logger.Err(err).Send()
						}
					}
				}
			}

			retMsg.Text = fmt.Sprintf("delete monitor ok! \ntoken address: %s\n", tokenAddr)
			_, err = tg.Send(retMsg)
			if err == nil {
				step.Clear()
			}
		}

	default:
		step.Clear()
		retMsg.Text = warnBedStep
		retMsg.ReplyMarkup = mainMenu
		tg.Send(retMsg)
	}
	err = tg.db.SaveStepToDb(chatIdStr, step)
	if err != nil {
		log.Logger.Err(err).Caller().Send()
	}
}

func IsHexAddress(s string) bool {
	if has0xPrefix(s) {
		s = s[2:]
	}
	return len(s) == 2*20 && isHex(s)
}

// has0xPrefix validates str begins with '0x' or '0X'.
func has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

// isHexCharacter returns bool of c being a valid hexadecimal.
func isHexCharacter(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

// isHex validates whether each byte is valid hexadecimal string.
func isHex(str string) bool {
	if len(str)%2 != 0 {
		return false
	}
	for _, c := range []byte(str) {
		if !isHexCharacter(c) {
			return false
		}
	}
	return true
}
