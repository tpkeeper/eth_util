package db

import (
	"encoding/json"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"math/big"
)

type MonitorTargetErc20 struct {
	ContractAddress string
	TokenAddress    string
	Amount          BigInt
	ChatId          map[string]struct{} //key is subscriber Id
}

type BigInt struct {
	big.Int
}

type Stack []string

func (s *Stack) Pop() (string, error) {
	if len(*s) == 0 {
		return "", fmt.Errorf("empty")
	}
	r := (*s)[len(*s)-1]
	*s = (*s)[0 : len(*s)-1]
	return r, nil
}

func (s *Stack) Push(e string) {
	*s = append(*s, e)
}

func (s *Stack) Top() string {
	if len(*s) == 0 {
		return ""
	}
	return (*s)[len(*s)-1]
}

func (s *Stack) Clear() {
	*s = (*s)[:0]
}

func (b BigInt) MarshalJSON() ([]byte, error) {
	return []byte(b.String()), nil
}

func (b *BigInt) UnmarshalJSON(p []byte) error {
	if string(p) == "null" {
		return nil
	}
	var z big.Int
	_, ok := z.SetString(string(p), 10)
	if !ok {
		return fmt.Errorf("not a valid big integer: %s", p)
	}
	b.Int = z
	return nil
}

func (db *Db) GetMonitorTargetErc20sFromDb() (map[string]*MonitorTargetErc20, error) {
	monitorTargetErc20s := make(map[string]*MonitorTargetErc20)
	err := (*bolt.DB)(db).View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(MonitorTargetErc20Bucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			monitorTargetErc20 := MonitorTargetErc20{}
			err := json.Unmarshal(v, &monitorTargetErc20)
			if err != nil {
				return fmt.Errorf("db unmarshal: %s", err)
			}
			monitorTargetErc20s[string(k)] = &monitorTargetErc20
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return monitorTargetErc20s, nil
}

func (db *Db) SaveMonitorTargetErc20sToDb(monitorTargetErc20s map[string]*MonitorTargetErc20) error {
	err := (*bolt.DB)(db).Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(MonitorTargetErc20Bucket))
		for key, monitorTargetErc20 := range monitorTargetErc20s {
			jbts, err := json.Marshal(monitorTargetErc20)
			if err != nil {
				return fmt.Errorf("json marshal: %s", err)
			}
			err = b.Put([]byte(key), jbts)
			if err != nil {
				return fmt.Errorf("bucket put: %s", err)
			}
		}
		return nil
	})
	return err
}

func (db *Db) SaveMonitorTargetErc20ToDb(monitorTargetErc20 MonitorTargetErc20) error {
	err := (*bolt.DB)(db).Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(MonitorTargetErc20Bucket))
		key := monitorTargetErc20.ContractAddress + monitorTargetErc20.TokenAddress
		jbts, err := json.Marshal(monitorTargetErc20)
		if err != nil {
			return fmt.Errorf("json marshal: %s", err)
		}
		err = b.Put([]byte(key), jbts)
		if err != nil {
			return fmt.Errorf("bucket put: %s", err)
		}
		return nil
	})
	return err
}

func (db *Db) DelMonitorTargetErc20FromDb(key string) error {
	err := (*bolt.DB)(db).Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(MonitorTargetErc20Bucket))

		err := b.Delete([]byte(key))
		if err != nil {
			return fmt.Errorf("bucket save: %s", err)
		}

		return nil
	})
	return err
}

func (db *Db) SaveStepToDb(chatId string, step Stack) error {
	err := (*bolt.DB)(db).Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TelegramBotStepBucket))
		jbts, err := json.Marshal(step)
		if err != nil {
			return fmt.Errorf("json marshal: %s", err)
		}
		err = b.Put([]byte(chatId), jbts)
		if err != nil {
			return fmt.Errorf("bucket put: %s", err)
		}

		return nil
	})
	return err
}

func (db *Db) GetStepFromDb(chatId string) (Stack, error) {
	var step Stack
	err := (*bolt.DB)(db).View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TelegramBotStepBucket))
		bts := b.Get([]byte(chatId)) //return nil if empty
		err := json.Unmarshal(bts, &step)
		if err != nil && len(bts) != 0 {
			return fmt.Errorf("db unmarshal: %s", err)
		}

		return nil
	})

	return step, err
}

func (db *Db) SaveTempContractAddrToDb(chatId string, addr string) error {
	err := (*bolt.DB)(db).Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TelegramBotTempContractBucket))
		jbts, err := json.Marshal(addr)
		if err != nil {
			return fmt.Errorf("json marshal: %s", err)
		}
		err = b.Put([]byte(chatId), jbts)
		if err != nil {
			return fmt.Errorf("bucket put: %s", err)
		}
		return nil
	})
	return err
}

func (db *Db) GetTempContractAddrFromDb(chatId string) (string, error) {
	var step string
	err := (*bolt.DB)(db).View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TelegramBotTempContractBucket))
		bts := b.Get([]byte(chatId)) //return nil if empty
		err := json.Unmarshal(bts, &step)
		if err != nil && len(bts) != 0 {
			return fmt.Errorf("db unmarshal: %s", err)
		}
		return nil
	})

	return step, err
}
