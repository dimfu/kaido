package store

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestOpenStore(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("error while getting working directory: %v\n", err)
	}
	parentDir := filepath.Dir(dir)
	outDir := path.Join(parentDir, "/.out")
	_, err = os.Stat(outDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(outDir, os.ModePerm)
		if err != nil {
			t.Fatalf("error while creating out directory: %v\n", err)
		}
	}

	storePath := path.Join(outDir, "kaido.store")
	store, err := Open(storePath)
	if err != nil {
		t.Fatalf("error while initializing store file: %v\n", err)
	}
	defer store.Close()
}

func TestWrite(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("error while getting working directory: %v\n", err)
	}
	parentDir := filepath.Dir(dir)
	outDir := path.Join(parentDir, "/.out")
	storePath := path.Join(outDir, "kaido.store")
	store, err := Open(storePath)
	if err != nil {
		t.Fatalf("error while initializing store file: %v\n", err)
	}
	defer store.Close()

	key := []byte("cool")
	value := []byte("cool values, such values")

	err = store.Put(Record{
		Key:   key,
		Value: value,
	})

	if err != nil {
		t.Fatalf("error while putting new record in the store: %v\n", err)
	}

	rec, err := store.Get("cool")
	if err != nil {
		t.Fatalf("error getting record for this key")
	}

	if rec == nil {
		t.Fatal("record should not be nil")
	}
}
