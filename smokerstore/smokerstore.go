package smokerstore

import (
	"log"
	"time"
	"encoding/binary"
	badger "github.com/dgraph-io/badger/v3"
)

type SmokerStore struct {
	Smokers map[string]time.Time
}

func New() *SmokerStore {
	ss := &SmokerStore{}
	ss.Smokers = make(map[string]time.Time)
	return ss
}

func (ss *SmokerStore) AddSmoker(email_addr string) {
	opts := badger.DefaultOptions("/tmp/smokers")
	opts.Logger = nil
	db, err := badger.Open(opts)
	handle(err)
	defer db.Close()
	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(email_addr), encode(time.Now()))
		return err
	})
	handle(err)
}

func (ss *SmokerStore) RegisterSmoker(email_addr string) {
	opts := badger.DefaultOptions("/tmp/smokers")
	opts.Logger = nil
	db, err := badger.Open(opts)
	handle(err)
	defer db.Close()
	err = db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(email_addr))
		return err
	})
	if err != nil {
		log.Printf("%s is ready for registration", email_addr)
	} else {
		log.Printf("%s has already registered", email_addr)
	}
}

func (ss *SmokerStore) DeleteSmoker(email_addr string) {
	opts := badger.DefaultOptions("/tmp/smokers")
	opts.Logger = nil
	db, err := badger.Open(opts)
	handle(err)
	defer db.Close()
	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(email_addr))
		return err
	})
	handle(err)
}

func (ss *SmokerStore) GetSmokers() {
	opts := badger.DefaultOptions("/tmp/smokers")
	opts.Logger = nil
	db, err := badger.Open(opts)
	handle(err)
	defer db.Close()
	err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func (v []byte) error {
				since := decode(v)
				log.Printf("key=%s, value=%s\n", k, since.Format(time.UnixDate))
				ss.Smokers[string(k)] = since
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	handle(err)
}

// decode unmarshals a time.
func decode(b []byte) time.Time {
    i := int64(binary.BigEndian.Uint64(b))
    return time.Unix(i, 0)
}

// encode marshals a time.
func encode(t time.Time) []byte {
    buf := make([]byte, 8)
    u := uint64(t.Unix())
    binary.BigEndian.PutUint64(buf, u)
    return buf
}

func handle(e error) {
	if e != nil {
		panic(e)
	}
}
