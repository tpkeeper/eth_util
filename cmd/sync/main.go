package main

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"golang.org/x/net/context"
)

type config struct {
	InfuraAPI string
	Mnemonic  string
	Schedule  string
	To        []string
}

var bigZero = big.NewInt(0)

func main() {
	cfg := config{}
	_, err := toml.DecodeFile("./conf.toml", &cfg)
	if err != nil {
		panic(err)
	}
	fmt.Printf("config: %s \n", cfg)

	_, err = ethclient.Dial(cfg.InfuraAPI)
	if err != nil {
		log.Panic(err)
	}

	wallet, err := hdwallet.NewFromMnemonic(cfg.Mnemonic)
	if err != nil {
		panic(err)
	}

	_, err = wallet.Derive(hdwallet.DefaultBaseDerivationPath, true)
	if err != nil {
		panic(err)
	}

	var addrTo []common.Address
	for _, to := range cfg.To {
		if !common.IsHexAddress(to) {
			panic(fmt.Sprintf("%s is not hex address", to))
		}
		addrTo = append(addrTo, common.HexToAddress(to))
	}

	ctrPairAbi, err := abi.JSON(strings.NewReader(pairABI))
	if err != nil {
		panic(err)
	}
	data, err := ctrPairAbi.Pack("sync")
	if err != nil {
		panic(err)
	}

	context := context.Background()
	//c := cron.New()
	//c.AddFunc(cfg.Schedule, func() {
	for index, _ := range addrTo {
		client, err := ethclient.Dial(cfg.InfuraAPI)
		if err != nil {
			fmt.Printf("connect infura err %e", err)
			return
		}

		for {
			err = sendContractTx(context, client, wallet, &addrTo[index], data)
			if err != nil {
				fmt.Printf("send tx for sync() err %s\n", err)
				fmt.Printf("will resend after 3 seconds \n")
				time.Sleep(time.Second*3)
				continue
			}
			break
		}

		fmt.Printf("sent tx for sync() ok, from:%s toContract:%s,time:%s\n\n",
			wallet.Accounts()[0].Address.String(), addrTo[index].String(), time.Now().String())
	}
	//fmt.Printf("schedule is running...\n")

	//})
	//c.Start()
	//fmt.Printf("schedule is running...\n")
	//select {}
}

func sendContractTx(ctx context.Context, client *ethclient.Client, wallet *hdwallet.Wallet, to *common.Address,
	data []byte) error {
	if len(wallet.Accounts()) == 0 {
		return errors.New("wallet have no address")
	}

	account := wallet.Accounts()[0]
	from := account.Address

	nonceNext, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		return err
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}
	gasPrice = new(big.Int).Add(gasPrice, big.NewInt(20e9))

	msg := ethereum.CallMsg{From: from, To: to, GasPrice: gasPrice, Value: bigZero, Data: data}

	gasNum, err := client.EstimateGas(ctx, msg)
	if err != nil {
		return err
	}
	gasNum = gasNum * 2

	tx := types.NewTransaction(nonceNext, *to,
		bigZero, gasNum, gasPrice, data)

	signedTx, err := wallet.SignTx(account, tx, nil)
	if err != nil {
		return err
	}
	signedTx.Hash()
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		return err
	}
	fmt.Printf("send tx hash: %s nonce: %d gasprice: %d gasLimit: %d\n", signedTx.Hash().String(),
		nonceNext, new(big.Int).Div(gasPrice, big.NewInt(1e9)).Int64(), gasNum)
	return nil
}
