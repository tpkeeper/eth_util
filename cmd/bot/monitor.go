package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

type TokenBalanceRes struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

func monitor(ctx context.Context, bot *tgbotapi.BotAPI) {
	ticker := time.NewTicker(time.Second * 10)
	api := "https://api.etherscan.io/api?module=account&action=tokenbalance&contractaddress=%s&address=%s&tag=latest&apikey=%s"
	defer ticker.Stop()
	//init amount of each monitorTargetErc20
	for _, monitorTargetErc20 := range monitorTargetErc20s {
		res, err := http.Get(fmt.Sprintf(api, monitorTargetErc20.ContractAddress,
			monitorTargetErc20.TokenAddress, "VC35I1VEW49ZTRNPDT11QQ8WWCS324FZGS"))

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
		fmt.Println("request", tokenBalanceRes)
		if tokenBalanceRes.Status != "1" {
			fmt.Println(tokenBalanceRes)
			continue
		}
		nowAmount := new(big.Int)
		nowAmount.SetString(tokenBalanceRes.Result, 10)
		nowAmount.Div(nowAmount, big.NewInt(1000000000000000000))

		monitorTargetErc20.Amount = *nowAmount
	}

	for {
		select {
		case <-ticker.C:
			for _, monitorTargetErc20 := range monitorTargetErc20s {
				res, err := http.Get(fmt.Sprintf(api, monitorTargetErc20.ContractAddress,
					monitorTargetErc20.TokenAddress, "VC35I1VEW49ZTRNPDT11QQ8WWCS324FZGS"))

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
				fmt.Println("request", tokenBalanceRes)
				if tokenBalanceRes.Status != "1" {
					fmt.Println(tokenBalanceRes)
					continue
				}
				nowAmount := new(big.Int)
				nowAmount.SetString(tokenBalanceRes.Result, 10)
				nowAmount.Div(nowAmount, big.NewInt(1000000000000000000))

				preAmount := monitorTargetErc20.Amount

				monitorTargetErc20.Amount = *nowAmount

				delta := new(big.Int).Sub(nowAmount, &preAmount)

				if delta.Cmp(big.NewInt(0)) != 0 {

					for chatId, _ := range monitorTargetErc20.ChatId {
						chatIdInt, err := strconv.ParseInt(chatId, 10, 64)
						if err != nil {
							fmt.Println(err)
							continue
						}
						msg := tgbotapi.NewMessage(chatIdInt,
							fmt.Sprintf("ContractAddress: %s\ntokenAddress: %s\nnowAmount: %s\ndelta: %s",
								monitorTargetErc20.ContractAddress, monitorTargetErc20.TokenAddress, nowAmount.String(), delta.String()))
						fmt.Println("bot send", msg)
						_, err = bot.Send(msg)
						if err != nil {
							fmt.Println(err)
							continue
						}
					}

				}

			}
		case <-ctx.Done():
			return
		}
	}
}
