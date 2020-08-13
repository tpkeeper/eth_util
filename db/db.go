package db

import (
	"fmt"
	bolt "go.etcd.io/bbolt"
)

const MonitorTargetErc20Bucket = "MonitorTargetErc20Bucket"

type Db bolt.DB


func NewDb(path string) (*Db, error) {
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

