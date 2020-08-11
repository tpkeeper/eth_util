package main

import (
	"encoding/json"
	"fmt"
	bolt "go.etcd.io/bbolt"
)

const MonitorTargetErc20Bucket = "MonitorTargetErc20Bucket"

func initDb(path string) error {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {

		_, err := tx.CreateBucketIfNotExists([]byte(MonitorTargetErc20Bucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}

func getMonitorTargetFromDb(path string) (map[string]*MonitorTargetErc20, error) {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	monitorTargetErc20s := make(map[string]*MonitorTargetErc20)
	err = db.View(func(tx *bolt.Tx) error {
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

func saveMonitorTargetToDb(path string, monitorTargetErc20s map[string]*MonitorTargetErc20) error {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(MonitorTargetErc20Bucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

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
