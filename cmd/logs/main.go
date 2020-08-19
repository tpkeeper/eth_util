package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
)

func main() {
	client, err := ethclient.Dial("wss://mainnet.infura.io/ws/v3/fa2b9f06ae2f43faa3ce69c80fde51c5")
	if err != nil {
		log.Fatal(err)
	}

	recept, err := client.TransactionReceipt(context.Background(), common.HexToHash("0x94bdc93ab71ade29917e5376cead5e80742aecbfa615c686f823af68229818fa"))
	if err != nil {
		log.Fatal(err)
	}
	for _, log := range recept.Logs {
		if log.Topics[0].String()=="0x7fcf532c15f0a6db0bd6d0e038bea71d30d808c7d98cb3bf7268a95bf5081b65"{
			amount:=new(big.Int).SetBytes(log.Data)
			fmt.Println(amount.Int64())
		}
	}
}
