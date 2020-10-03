package main

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
)

func main() {
	client, err := ethclient.Dial("wss://reposten.infura.io/ws/v3/fa2b9f06ae2f43faa3ce69c80fde51c5")
	if err != nil {
		log.Fatal(err)
	}

	//recept, err := client.TransactionReceipt(context.Background(), common.HexToHash("0x94bdc93ab71ade29917e5376cead5e80742aecbfa615c686f823af68229818fa"))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//for _, log := range recept.Logs {
	//	if log.Topics[0].String()=="0x7fcf532c15f0a6db0bd6d0e038bea71d30d808c7d98cb3bf7268a95bf5081b65"{
	//		amount:=new(big.Int).SetBytes(log.Data)
	//		fmt.Println(amount.Int64())
	//	}
	//}
	filterQuery := ethereum.FilterQuery{
		FromBlock: big.NewInt(8738755),
		Addresses: []common.Address{common.HexToAddress("0x66570e54591ed0372a139529726b4c7f44da7c74")},
		Topics: [][]common.Hash{common.HexToHash("0xb94bf7f9302edf52a596286915a69b4b0685574cffdedd0712e3c62f2550f0ba")}}
	logs, err := client.FilterLogs(context.Background(), filterQuery)
}
