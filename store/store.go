package store

import (
	"bytes"
	"encoding/gob"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username  string
	Password  []byte
	Active    bool
	StoppedAt time.Time
}

type Store struct {
	Users map[string]User
}

func New() *Store {
	store := &Store{}
	store.Users = make(map[string]User)
	return store
}

func (store *Store) AddUser(emailAddr string, username string, password string) {
	opts := badger.DefaultOptions("C:/Users/Jakob/roundbyte/smokers")
	opts.Logger = nil
	db, err := badger.Open(opts)
	handle(err)
	defer db.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	userEncode := User{Username: username, Password: hashedPassword, Active: true, StoppedAt: time.Now()}
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	if err = e.Encode(userEncode); err != nil {
		panic(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(emailAddr), b.Bytes())
		return err
	})
	handle(err)
}

func (store *Store) CheckUserPassword(emailAddr string, password string) bool {
	opts := badger.DefaultOptions("C:/Users/Jakob/roundbyte/smokers")
	opts.Logger = nil
	db, err := badger.Open(opts)
	handle(err)
	defer db.Close()

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(emailAddr))
		handle(err)
		val, err := item.ValueCopy(nil)
		var userDecode User
		d := gob.NewDecoder(bytes.NewReader(val))
		err = d.Decode(&userDecode)
		handle(err)
		if err = bcrypt.CompareHashAndPassword(userDecode.Password, []byte(password)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return false
	}
	return true
}

func (s *Store) GetAllUsers() {
	opts := badger.DefaultOptions("C:/Users/Jakob/roundbyte/smokers")
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
			key := item.Key()
			err := item.Value(func(val []byte) error {
				var userDecode User
				d := gob.NewDecoder(bytes.NewReader(val))
				err := d.Decode(&userDecode)
				handle(err)
				s.Users[string(key)] = userDecode
				return nil
			})
			handle(err)
		}
		return nil
	})
	handle(err)
}

func handle(e error) {
	if e != nil {
		panic(e)
	}
}
