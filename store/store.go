package store

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const (
	HEADER_SIZE = 12 // Timestamp = 4 bytes, Key = 4 bytes, Value = 4 bytes
)

type Record struct {
	Timestamp uint32
	Key       []byte
	Value     []byte
}

type Store struct {
	storage map[string]int64
	mu      sync.RWMutex
	file    *os.File
}

var (
	mu       = sync.Mutex{}
	instance *Store
	once     sync.Once
)

func GetInstance() (*Store, error) {
	if instance == nil {
		mu.Lock()
		defer mu.Unlock()
		// more if check for instance to ensure no more than 1 goroutine bypass the first check
		if instance == nil {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, err
			}
			dbDir := fmt.Sprintf("%s/.kaido/store.db", homeDir)
			store, err := open(dbDir)
			if err != nil {
				return nil, err
			}
			instance = store
		}
	}
	return instance, nil
}

func open(path string) (*Store, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("error while open file: %v\n", err)
	}

	store := &Store{
		storage: make(map[string]int64),
		mu:      sync.RWMutex{},
		file:    file,
	}

	if err := store.generateIndex(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file != nil {
		defer s.file.Close()
		err := s.compact()
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func (s *Store) Get(key string) (*Record, error) {
	offset, exists := s.storage[key]
	if !exists {
		return nil, errors.New("could not find record with this key")
	}

	rec, err := s.deserialize(offset)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

func (s *Store) Put(r Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// gets the offset of the new record
	offset, err := s.file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	// writes new value at the very end
	if _, err = s.file.Write(recTobuf(r)); err != nil {
		return err
	}

	s.storage[string(r.Key)] = offset
	return nil
}

func (s *Store) generateIndex() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var offset int64
	for {
		rec, err := s.deserialize(offset)
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		s.storage[string(rec.Key)] = offset
		offset += int64(HEADER_SIZE + len(rec.Key) + len(rec.Value))
	}

	return nil
}

func (s *Store) compact() error {
	storePath, err := filepath.Abs(s.file.Name())
	if err != nil {
		return err
	}

	dir := filepath.Dir(storePath)

	temp, err := os.CreateTemp(dir, "temp-*.db")
	if err != nil {
		return err
	}
	defer temp.Close()

	_, err = temp.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	// write recent records from store to temp file
	for key := range s.storage {
		record, err := s.Get(key)
		if err != nil {
			continue
		}

		if _, err = temp.Write(recTobuf(*record)); err != nil {
			return err
		}
	}

	err = os.Remove(storePath)
	if err != nil {
		return err
	}

	err = os.Rename(temp.Name(), s.file.Name())
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) deserialize(offset int64) (*Record, error) {
	if _, err := s.file.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}

	record := &Record{}
	var timestamp, keyLen, valueLen uint32

	if err := binary.Read(s.file, binary.LittleEndian, &timestamp); err != nil {
		return nil, err
	}
	record.Timestamp = timestamp

	if err := binary.Read(s.file, binary.LittleEndian, &keyLen); err != nil {
		return nil, err
	}

	if err := binary.Read(s.file, binary.LittleEndian, &valueLen); err != nil {
		return nil, err
	}

	key := make([]byte, keyLen)
	if _, err := s.file.Read(key); err != nil {
		return nil, err
	}
	record.Key = key

	value := make([]byte, valueLen)
	if _, err := s.file.Read(value); err != nil {
		return nil, err
	}
	record.Value = value

	return record, nil
}

func recTobuf(r Record) []byte {
	size := HEADER_SIZE + len(r.Key) + len(r.Value)
	buf := make([]byte, size)

	// serialize the header for this record
	binary.LittleEndian.PutUint32(buf[0:4], r.Timestamp)
	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(r.Key)))
	binary.LittleEndian.PutUint32(buf[8:12], uint32(len(r.Value)))

	copy(buf[HEADER_SIZE:HEADER_SIZE+len(r.Key)], r.Key)
	copy(buf[HEADER_SIZE+len(r.Key):], r.Value)
	return buf
}
