package main

import (
	"encoding/json"

	"github.com/boltdb/bolt"
)

const AllIds = "ALL_IDS_KEY"

type StorageService interface {
	Results(id string) *jstats
	SaveResults(id string, j *jstats) error
	AllIds() []string
}

type BoltStorageService struct {
	DB         *bolt.DB
	BucketName []byte
}

func (s *BoltStorageService) Results(id string) (*jstats, error) {
	var j jstats
	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.BucketName))
		v := b.Get([]byte(id))
		if err := unmarshal(v, &j); err != nil {
			return err
		}
		return nil
	})
	return &j, err
}

func (s *BoltStorageService) SaveResults(j *jstats) error {
	tx, err := s.DB.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists([]byte(s.BucketName))
	if err != nil {
		return err
	}

	// Marshal and insert record.
	if jstatsJson, err := marshal(j); err != nil {
		return err
	} else if err := b.Put([]byte(j.Key()), jstatsJson); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *BoltStorageService) AllIds() ([]string, error) {
	r := []string{}
	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.BucketName))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			// fmt.Printf("key=%s, value=%s\n", k, v)
			r = append(r, string(k))
		}
		return nil
	})
	return r, err
}

func marshal(j *jstats) ([]byte, error) {
	return json.Marshal(j)
}

func unmarshal(d []byte, j *jstats) error {
	return json.Unmarshal(d, &j)
}
