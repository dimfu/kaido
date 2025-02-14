package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	bolt "go.etcd.io/bbolt"
)

func init() {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	storePath := path.Join(dir, ".kaido")
	_, err = os.Stat(storePath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(storePath, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	dbDir := fmt.Sprintf("%s/.kaido/store.db", homeDir)
	db, err := bolt.Open(dbDir, 0600, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
