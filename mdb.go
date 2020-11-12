package mdb

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"sort"
	"sync"
)

var (
	// DefaultConfig .
	//   Path == "" // run in ram
	//
	DefaultConfig = defaultConfig()
)

type (
	// Config holds the db configuration:
	//   Path is the filepath of the db file
	Config struct {
		// Path if this is an empty string,
		// then its stored in ram,
		// otherwise its the filepath
		Path string
	}
	// DB .
	//  just a simple map of maps of strings that can persist to disk
	DB struct {
		db     map[string]map[string]string // the db data
		config Config						// db config
		file   *os.File						// db file
		sync.RWMutex						// sync lock
	}
	// KV is just a key-value pair
	KV struct {
		Key, Value string
	}
)

// Open opens an existing database file, or a new path.
// open nil to receive the default config
func Open(config *Config) (*DB, func() error, error) {
	db := &DB{}
	db.db = map[string]map[string]string{}
	if config == nil {
		db.config = DefaultConfig
	} else {
		db.config = *config
	}
	if db.config.Path != "" {
		err := *new(error)
		db.file, err = os.OpenFile(db.config.Path, os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			return nil, nil, err
		}
		err = db.load()
		if err != nil {
			db.file.Close()
			return nil, nil, err
		}
	}
	return db,
		func() error {
			err := db.Save()
			if err != nil {
				return err
			}
			db.Lock()
			defer db.Unlock()
			if db.config.Path != "" {
				err = db.file.Sync()
				if err != nil {
					return err
				}
				err = db.file.Close()
				if err != nil {
					return err
				}
			}
			db = &DB{db: map[string]map[string]string{}}
			return nil
		},
		nil
}

// Save .
func (db *DB) Save() error {
	if db.config.Path != "" {
		db.RLock()
		defer db.RUnlock()
		b := bytes.Buffer{}
		enc := gob.NewEncoder(&b)
		err := enc.Encode(db.db)
		if err != nil {
			return err
		}
		// wtf am i doing here? TODO
		_, err = db.file.WriteAt(b.Bytes(), 0)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close .
func (db *DB) Close() error {
	err := db.Save()
	if err != nil {
		return err
	}
	db.Lock()
	defer db.Unlock()
	if db.config.Path != "" {
		err = db.file.Sync()
		if err != nil {
			return err
		}
		err = db.file.Close()
		if err != nil {
			return err
		}
	}
	db = &DB{db: map[string]map[string]string{}}
	return nil
}

// GetMap .
func (db *DB) GetMap(s string) (map[string]string, error) {
	db.RLock()
	defer db.RUnlock()
	if (len(db.db[s]) == 0 || db.db[s] == nil) {
		return map[string]string{}, fmt.Errorf("%s", "missing")
	}
	return db.db[s], nil
}

// GetKV .
func (db *DB) GetKV(s string) ([]KV, error) {
	db.RLock()
	defer db.RUnlock()
	if (len(db.db[s]) == 0 || db.db[s] == nil) {
		return []KV{}, fmt.Errorf("%s", "missing")
	}
	kv := []KV{}
	for k, v := range db.db[s] {
		kv = append(kv, KV{k, v})
	}
	sort.SliceStable(
		kv,
		func(i, j int) bool {
			return kv[i].Key < kv[j].Key
		},
	)
	return kv, nil
}

// SetMap .
func (db *DB) SetMap(s string, m map[string]string) error {
	db.Lock()
	defer db.Unlock()
	for k, v := range m {
		if db.db[s] == nil {
			db.db[s] = map[string]string{}
		}
		db.db[s][k] = v
	}
	return nil
}

// SetKV .
func (db *DB) SetKV(s string, kv []KV) error {
	db.Lock()
	defer db.Unlock()
	m := map[string]string{}
	for _, p := range kv {
		if _, present := m[p.Key]; !present {
			m[p.Key] = p.Value
		} else {
			return fmt.Errorf("%s", "duplicate key detected")
		}
	}
	for k, v := range m {
		if db.db[s] == nil {
			db.db[s] = map[string]string{}
		}
		db.db[s][k] = v
	}
	return nil
}

// Delete .
func (db *DB) Delete(s string) error {
	delete(db.db, s)
	return nil
}

func (db *DB) load() error {
	fi, err := db.file.Stat()
	if err != nil {
		return err
	}
	if fi.Size() == 0 {
		err = db.init()
		if err != nil {
			return err
		}
	}
	err = gob.NewDecoder(db.file).Decode(&db.db)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) init() error {
	db.Lock()
	defer db.Unlock()
	db.db = map[string]map[string]string{}
	b := bytes.Buffer{}
	err := gob.NewEncoder(&b).Encode(db.db)
	if err != nil {
		return err
	}
	// wtf am i doing here? TODO
	_, err = db.file.WriteAt(b.Bytes(), 0)
	if err != nil {
		return err
	}
	err = db.file.Sync()
	if err != nil {
		return err
	}
	return nil
}

func defaultConfig() Config {
	return Config{
		Path: "",
	}
}
