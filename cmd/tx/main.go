package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/miguelmota/go-ethereum-hdwallet"
	"golang.org/x/net/context"
	"log"
	"math/big"
	"time"
)

type config struct {
	Api      string
	Mnemonic string
	Value    float64 //eth
	GasLimit uint64
	GasPrice uint64 //Gwei
	Times    uint64
	Interval uint64 //ms
}

func main() {
	cfg := config{}
	_, err := toml.DecodeFile("./conf.toml", &cfg)
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg)
	client, err := ethclient.Dial(cfg.Api)
	if err != nil {
		log.Fatal(err)
	}
	wallet, err := hdwallet.NewFromMnemonic(cfg.Mnemonic)
	if err != nil {
		panic(err)
	}
	account0, err := wallet.Derive(hdwallet.DefaultBaseDerivationPath, true)
	if err != nil {
		panic(err)
	}

	wallet.Accounts()

	context := context.Background()
	base := new(big.Float).SetInt(big.NewInt(1000000000000000000))
	valueFloat := new(big.Float).Mul(big.NewFloat(cfg.Value), base)
	valueInt := new(big.Int)
	valueFloat.Int(valueInt)

	gasPriceWei := new(big.Int).Mul(big.NewInt(int64(cfg.GasPrice)), big.NewInt(1000000000))

	duration := time.Duration(cfg.Interval)
	ticker := time.NewTicker(time.Millisecond * duration)
	defer ticker.Stop()

	var i uint64
	for i = 0; i < cfg.Times; i++ {
		select {
		case <-ticker.C:
			nonceNext, err := client.PendingNonceAt(context, account0.Address)
			if err != nil {
				fmt.Println(err)
			}
			tx := types.NewTransaction(nonceNext, account0.Address,
				valueInt, cfg.GasLimit, gasPriceWei, nil)
			signedTx, err := wallet.SignTx(account0, tx, nil)
			if err != nil {
				fmt.Println(err)
			}
			err = client.SendTransaction(context, signedTx)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("send tx , nonce %d \n", nonceNext)

		}

	}
}
