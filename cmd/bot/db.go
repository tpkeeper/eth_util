package main

import (
	"encoding/json"
	"fmt"
	bolt "go.etcd.io/bbolt"
)

const MonitorTargetErc20Bucket = "MonitorTargetErc20Bucket"

type Db bolt.DB

func newDb(path string) (*Db, error) {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {

		_, err := tx.CreateBucketIfNotExists([]byte(MonitorTargetErc20Bucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return (*Db)(db), nil

}

func (db *Db) Close() error {
	return (*bolt.DB)(db).Close()
}

func (db *Db) GetMonitorTargetErc20sFromDb() (map[string]*MonitorTargetErc20, error) {
	monitorTargetErc20s := make(map[string]*MonitorTargetErc20)
	err := (*bolt.DB)(db).View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
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
