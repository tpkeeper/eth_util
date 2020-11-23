package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/tpkeeper/eth-util/log"

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

type RootHash [32]byte

type ResClaimRoot struct {
	Data string `json:"data"`
	Ok   bool   `json:"ok"`
}

func main() {
	cfg := config{}
	_, err := toml.DecodeFile("./conf.toml", &cfg)
	if err != nil {
		panic(err)
	}
	log.InitLogFile("./log_data")
	logrus.Printf("config: %s \n", cfg)

	_, err = ethclient.Dial(cfg.InfuraAPI)
	if err != nil {
		panic(err)
	}

	//wallet, err := hdwallet.NewFromMnemonic(cfg.Mnemonic)
	//if err != nil {
	//	panic(err)
	//}
	seedBts, err := hex.DecodeString(cfg.Mnemonic)
	if err != nil {
		panic(err)
	}
	wallet, err := hdwallet.NewFromSeed(seedBts)
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

	claimAbi, err := abi.JSON(strings.NewReader(abiClaim))
	if err != nil {
		panic(err)
	}
	pauseMethodData, err := claimAbi.Pack("pause")
	if err != nil {
		panic(err)
	}

	merkleRootMethodData, err := claimAbi.Pack("merkleRoot")
	if err != err {
		panic(err)
	}
	startMethodData, err := claimAbi.Pack("start")
	if err != err {
		panic(err)
	}
	var testRootHash RootHash
	_, err = claimAbi.Pack("changeMerkleRoot", testRootHash)
	if err != err {
		panic(err)
	}

	context := context.Background()

	var oldMerkelRoot RootHash

	for {
		client, err := ethclient.Dial(cfg.InfuraAPI)
		if err != nil {
			logrus.Printf("connect infura err %e", err)
			return
		}
		logrus.Println("merkleRootmethod", hex.EncodeToString(merkleRootMethodData))
		resBytes, err := callContract(context, client, wallet, &addrTo[0], merkleRootMethodData)
		if err != nil {
			logrus.Printf("callContract  for merkleRoot err %s\n", err)
			logrus.Printf("will request after 5 seconds \n")
			time.Sleep(time.Second * 5)
			continue
		}
		err = claimAbi.Unpack(&oldMerkelRoot, "merkleRoot", resBytes)
		if err != nil {
			logrus.Printf("callContract  for merkleRoot unpack err：%s，resBytes %v\n", err, resBytes)
			logrus.Printf("will request after 5 seconds \n")
			time.Sleep(time.Second * 5)
			continue
		}
		logrus.Printf("get old merkleRoot ok. old merkleRoot %s from:%s toContract:%s,time:%s\n\n",
			hex.EncodeToString(oldMerkelRoot[:]), wallet.Accounts()[0].Address.String(), addrTo[0].String(), time.Now().String())
		break
	}

	for {
		client, err := ethclient.Dial(cfg.InfuraAPI)
		if err != nil {
			logrus.Printf("connect infura err %e", err)
			return
		}
		err = sendContractTx(context, client, wallet, &addrTo[0], pauseMethodData)
		if err != nil {
			logrus.Printf("send tx for pause() err %s\n", err)
			logrus.Printf("will resend after 5 seconds \n")
			time.Sleep(time.Second * 5)
			continue
		}

		logrus.Printf("sent tx for pause() ok, from:%s toContract:%s,time:%s\n\n",
			wallet.Accounts()[0].Address.String(), addrTo[0].String(), time.Now().String())

		break
	}

	for {
		res, err := http.Get("https://api.miniswap.org/api/mini/claim_root")
		var resClaimRoot ResClaimRoot
		bts, err := ioutil.ReadAll(res.Body)
		err = json.Unmarshal(bts, &resClaimRoot)
		if err != nil {
			logrus.Printf("get claim root err %s\n", err)
			logrus.Printf("will request after 5 seconds \n")
			time.Sleep(time.Second * 5)
			continue
		}

		if resClaimRoot.Data == hex.EncodeToString(oldMerkelRoot[:]) {
			logrus.Printf("api claim root not change \n")
			logrus.Printf("will request after 5 seconds \n")
			time.Sleep(time.Second * 5)
			continue
		}

		client, err := ethclient.Dial(cfg.InfuraAPI)
		if err != nil {
			logrus.Printf("connect infura err %e", err)
			continue
		}
		var rootHash RootHash
		rootBts, err := hex.DecodeString(resClaimRoot.Data)
		if err != nil {
			logrus.Printf("root hash decode err %e", err)
			continue
		}

		copy(rootHash[:], rootBts[:])

		changeMerkleRootMethodData, err := claimAbi.Pack("changeMerkleRoot", rootHash)
		if err != err {
			panic(err)
		}
		err = sendContractTx(context, client, wallet, &addrTo[0], changeMerkleRootMethodData)
		if err != nil {
			logrus.Printf("send tx for changeMerkleRoot() err %s\n", err)
			logrus.Printf("will resend after 5 seconds \n")
			time.Sleep(time.Second * 5)
			continue
		}
		logrus.Printf("sent tx for changeMerkleRoot() ok,merkle root: %s from:%s toContract:%s,time:%s\n\n",
			resClaimRoot.Data, wallet.Accounts()[0].Address.String(), addrTo[0].String(), time.Now().String())
		break

	}

	for {
		client, err := ethclient.Dial(cfg.InfuraAPI)
		if err != nil {
			logrus.Printf("connect infura err %e", err)
			return
		}
		err = sendContractTx(context, client, wallet, &addrTo[0], startMethodData)
		if err != nil {
			logrus.Printf("send tx for start() err %s\n", err)
			logrus.Printf("will resend after 5 seconds \n")
			time.Sleep(time.Second * 5)
			continue
		}

		logrus.Printf("sent tx for start() ok, from:%s toContract:%s,time:%s\n\n",
			wallet.Accounts()[0].Address.String(), addrTo[0].String(), time.Now().String())
		break
	}

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
	gasNum = gasNum * 3 / 2

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
	logrus.Printf("send tx hash: %s nonce: %d gasprice: %d gasLimit: %d\n", signedTx.Hash().String(),
		nonceNext, new(big.Int).Div(gasPrice, big.NewInt(1e9)).Int64(), gasNum)
	return nil
}

func callContract(ctx context.Context, client *ethclient.Client, wallet *hdwallet.Wallet, to *common.Address, data []byte) ([]byte, error) {
	if len(wallet.Accounts()) == 0 {
		return nil, errors.New("wallet have no address")
	}
	account := wallet.Accounts()[0]
	from := account.Address

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}
	gasPrice = new(big.Int).Add(gasPrice, big.NewInt(20e9))

	msg := ethereum.CallMsg{From: from, To: to, GasPrice: gasPrice, Value: bigZero, Data: data}
	return client.CallContract(ctx, msg, nil)
}
