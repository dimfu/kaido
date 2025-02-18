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

var (
	ERR_KEY_NOT_FOUND = errors.New("could not find record with this key")
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
		return s.compact()
	}
	return nil
}

func (s *Store) Get(key string) (*Record, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("key cannot be empty")
	}

	s.mu.RLock()
	offset, exists := s.storage[key]
	s.mu.RUnlock()

	if !exists {
		return nil, ERR_KEY_NOT_FOUND
	}

	return s.deserialize(offset)
}

func (s *Store) Put(r Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// gets the offset of the new record
	offset, err := s.file.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("could not set offset: %v", err)
	}

	// writes new value at the very end
	if _, err = s.file.Write(recTobuf(r)); err != nil {
		return fmt.Errorf("error while writing new value: %v", err)
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
			return fmt.Errorf("error while generating index: %v", err)
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
	for _, offset := range s.storage {
		record, err := s.deserialize(offset)
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
	var timestamp, keyLen, valueLen uint32
	record := &Record{}
	buf := make([]byte, HEADER_SIZE)

	if _, err := s.file.ReadAt(buf, offset); err != nil {
		return nil, err
	}

	timestamp = binary.LittleEndian.Uint32(buf[0:4])
	keyLen = binary.LittleEndian.Uint32(buf[4:8])
	valueLen = binary.LittleEndian.Uint32(buf[8:12])

	record.Timestamp = timestamp

	key := make([]byte, keyLen)
	if _, err := s.file.ReadAt(key, HEADER_SIZE+offset); err != nil {
		return nil, err
	}
	record.Key = key

	value := make([]byte, valueLen)
	if _, err := s.file.ReadAt(value, HEADER_SIZE+offset+int64(keyLen)); err != nil {
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
