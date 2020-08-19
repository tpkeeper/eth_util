package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tpkeeper/eth-util/db"
	"github.com/tpkeeper/eth-util/log"
	"github.com/tpkeeper/eth-util/notify"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"
)

var apiPre = "https://apiPre.etherscan.io/apiPre?module=account&action=tokenbalance&contractaddress=%s&address=%s&tag=latest&apikey=%s"

type TokenBalanceRes struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

type Erc20Monitor struct {
	db       *db.Db
	notifier []notify.Notifier
}

func NewErc20Monitor(db *db.Db, notifier []notify.Notifier) *Erc20Monitor {
	return &Erc20Monitor{db, notifier}
}

func (m *Erc20Monitor) Start(ctx context.Context) {

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	monitorTargetErc20s, err := m.db.GetMonitorTargetErc20sFromDb()
	if err != nil {
		panic(err)
	}

	//init amount of each monitorTargetErc20
	for _, monitorTargetErc20 := range monitorTargetErc20s {
		api := fmt.Sprintf(apiPre, monitorTargetErc20.ContractAddress,
			monitorTargetErc20.TokenAddress, "VC35I1VEW49ZTRNPDT11QQ8WWCS324FZGS")
		res, err := http.Get(api)

		if err != nil {
			log.Logger.Err(err).Send()
			continue
		}
		tokenBalanceRes := TokenBalanceRes{}
		bts, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Logger.Err(err).Send()
			continue
		}
		err = json.Unmarshal(bts, &tokenBalanceRes)
		if err != nil {
			log.Logger.Err(err).Send()
			continue
		}
		if tokenBalanceRes.Status != "1" {
			log.Logger.Error().
				Str("status", tokenBalanceRes.Status).
				Str("result", tokenBalanceRes.Result).
				Str("msg", tokenBalanceRes.Message).
				Send()
			continue
		}
		nowAmount := new(big.Int)
		nowAmount.SetString(tokenBalanceRes.Result, 10)
		nowAmount.Div(nowAmount, big.NewInt(1000000000000000000))

		monitorTargetErc20.Amount = db.BigInt{*nowAmount}
	}

	m.db.SaveMonitorTargetErc20sToDb(monitorTargetErc20s)

	for {
		select {
		case <-ticker.C:

			monitorTargetErc20s, err := m.db.GetMonitorTargetErc20sFromDb()
			if err != nil {
				log.Logger.Err(err).Send()
				continue
			}
			for _, monitorTargetErc20 := range monitorTargetErc20s {
				api := fmt.Sprintf(apiPre, monitorTargetErc20.ContractAddress,
					monitorTargetErc20.TokenAddress, "VC35I1VEW49ZTRNPDT11QQ8WWCS324FZGS")
				res, err := http.Get(api)

				if err != nil {
					log.Logger.Err(err).Str("api", api).Send()
					continue
				}
				tokenBalanceRes := TokenBalanceRes{}
				bts, err := ioutil.ReadAll(res.Body)
				if err != nil {
					log.Logger.Err(err).Send()
					continue
				}
				err = json.Unmarshal(bts, &tokenBalanceRes)
				if err != nil {
					log.Logger.Err(err).Send()
					continue
				}
				log.Logger.Info().
					Str("api",api).
					Str("res",tokenBalanceRes.Result)

				if tokenBalanceRes.Status != "1" {
					log.Logger.Error().
						Str("status", tokenBalanceRes.Status).
						Str("result", tokenBalanceRes.Result).
						Str("msg", tokenBalanceRes.Message).
						Send()
					continue
				}
				nowAmount := new(big.Int)
				nowAmount.SetString(tokenBalanceRes.Result, 10)
				nowAmount.Div(nowAmount, big.NewInt(1000000000000000000))

				preAmount := monitorTargetErc20.Amount

				monitorTargetErc20.Amount = db.BigInt{*nowAmount}

				delta := new(big.Int).Sub(nowAmount, &preAmount.Int)

				if delta.Cmp(big.NewInt(0)) != 0 {
					err := m.db.SaveMonitorTargetErc20ToDb(*monitorTargetErc20)
					if err != nil {
						log.Logger.Err(err).Send()
					}
					for chatId, _ := range monitorTargetErc20.ChatId {
						msg := fmt.Sprintf("ContractAddress: %s\ntokenAddress: %s\nnowAmount: %s\ndelta: %s",
							monitorTargetErc20.ContractAddress, monitorTargetErc20.TokenAddress, nowAmount.String(), delta.String())
						for _, notifier := range m.notifier {
							err = notifier.Notify(chatId, msg)
							if err != nil {
								log.Logger.Err(err).Send()
								continue
							}
						}
					}

				}

			}
		case <-ctx.Done():
			return
		}
	}
}
