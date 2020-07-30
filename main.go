package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
)

type  IssueTo struct {
	To common.Address
	Amount *big.Int
}

func main() {
	clientHttp,err:=ethclient.Dial("https://ropsten.infura.io/v3/fa2b9f06ae2f43faa3ce69c80fde51c5")
	client, err := ethclient.Dial("wss://ropsten.infura.io/ws/v3/fa2b9f06ae2f43faa3ce69c80fde51c5")
	if err != nil {
		log.Fatal(err)
	}
	contractAddress := common.HexToAddress("0xbf004d22007118194f4578920f507AeE0EE11649")
	issueTopic:=common.HexToHash("0x95b20c4dce7884f76ac43c1deba4dc6c3f968fdbb364501891f11021bc17b6f2")
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
		Topics: [][]common.Hash{{issueTopic}},
		FromBlock: big.NewInt(8397502),
	}
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal(err)
	}

	contractAbi,err:=abi.JSON(strings.NewReader(abiMini))
	if err != nil {
		log.Fatal(err)
	}
	var issueTo IssueTo

	for {
		fmt.Println("start listen:")
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case vLog := <-logs:
			contractAbi.Unpack(&issueTo,"IssueTo",vLog.Data)
			fmt.Println(vLog.TxHash.String()) // pointer to event log
			fmt.Println(issueTo.To.String())
			fmt.Println(issueTo.Amount.Int64())
			tx,ispending,err:=clientHttp.TransactionByHash(context.Background(),vLog.TxHash)
			if err!=nil{
				fmt.Println(err)
			}
			fmt.Println(ispending,tx.Value())
			//fmt.Println(vLog.)
		}
	}
}
