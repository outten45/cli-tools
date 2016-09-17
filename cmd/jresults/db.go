package main

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
)

const AllIds = "ALL_IDS_KEY"

type StorageService interface {
	Results(id string) *jstats
	AllIds() []string
}

type BoltStorageService struct {
	DB         *bolt.DB
	BucketName []byte
}

func (s *BoltStorageService) Results(id string) (*jstats, error) {
	var j *jstats
	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.BucketName))
		v := b.Get([]byte(id))
		if err := unmarshal(v, j); err != nil {
			return err
		}
		return nil
	})
	return j, err
}

func (s *BoltStorageService) AllIds() ([]string, error) {
	r := []string{}
	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.BucketName))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
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
