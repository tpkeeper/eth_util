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
	ticker := time.NewTicker(time.Second * 5)
	api := "https://api.etherscan.io/api?module=account&action=tokenbalance&contractaddress=%s&address=%s&tag=latest&apikey=%s"
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			for _, tokenInfo := range monitorTargetErc20s {
				res, err := http.Get(fmt.Sprintf(api, tokenInfo.ContractAddress, tokenInfo.TokenAddress, "VC35I1VEW49ZTRNPDT11QQ8WWCS324FZGS"))
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
				nowAmount := new(big.Int)
				nowAmount.SetString(tokenBalanceRes.Result, 10)
				nowAmount.Div(nowAmount, big.NewInt(1000000000000000000))

				preAmount := tokenInfo.Amount

				tokenInfo.Amount = *nowAmount

				delta := new(big.Int).Sub(nowAmount, &preAmount)

				if delta.Cmp(big.NewInt(0)) != 0 {
					for chatId, _ := range tokenInfo.ChatId {
						chatIdInt, err := strconv.ParseInt(chatId, 10, 64)
						if err != nil {
							fmt.Println(err)
							continue
						}
						msg := tgbotapi.NewMessage(chatIdInt,
							fmt.Sprintf("ContractAddress: %s\ntokenAddress: %s\nnowAmount: %s\ndelta: %s",
								tokenInfo.ContractAddress, tokenInfo.TokenAddress, nowAmount.String(), delta.String()))
						_, err = bot.Send(msg)
						if err != nil {
							fmt.Println(err)
						}
					}
				}

			}
		case <-ctx.Done():
			return
		}
	}
}
