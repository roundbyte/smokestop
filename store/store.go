package store

import (
	"bytes"
	"encoding/gob"
	"log"
	"os"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	guuid "github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	EmailAddr string
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

func (store *Store) RegisterUser(emailAddr string, username string, password string) (string, error) {
	var err error
	var db *badger.DB

	opts := badger.DefaultOptions(os.Getenv("DBPATH"))
	opts.Logger = nil
	db, err = badger.Open(opts)
	if err != nil {
		return "", err
	}
	defer db.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return "", err
	}
	userEncode := User{
		EmailAddr: emailAddr,
		Username:  username,
		Password:  hashedPassword,
		Active:    true,
		StoppedAt: time.Now(),
	}

	var byteEncodedUser bytes.Buffer
	encoder := gob.NewEncoder(&byteEncodedUser)
	err = encoder.Encode(userEncode)
	if err != nil {
		return "", err
	}
	var id string
	err = db.Update(func(txn *badger.Txn) error {
		id = guuid.NewString()
		return txn.Set([]byte(id), byteEncodedUser.Bytes())
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (store *Store) DoesPasswordMatch(userId string, password string) error {
	var err error
	var db *badger.DB

	opts := badger.DefaultOptions(os.Getenv("DBPATH"))
	opts.Logger = nil
	db, err = badger.Open(opts)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(userId))
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		var decodedUser User
		d := gob.NewDecoder(bytes.NewReader(val))
		err = d.Decode(&decodedUser)
		if err != nil {
			return err
		}
		if err = bcrypt.CompareHashAndPassword(decodedUser.Password, []byte(password)); err != nil {
			return err
		}
		return nil
	})
}

func (store *Store) GetAllUsers() error {
	var err error
	var db *badger.DB

	opts := badger.DefaultOptions(os.Getenv("DBPATH"))
	opts.Logger = nil
	db, err = badger.Open(opts)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			log.Println(string(key))
			err := item.Value(func(val []byte) error {
				var decodedUser User
				d := gob.NewDecoder(bytes.NewReader(val))
				err := d.Decode(&decodedUser)
				if err != nil {
					return err
				}
				store.Users[string(key)] = decodedUser
				return nil
			})
			if err != nil {
				log.Println("Error with an entry")
			}
		}
		return nil
	})
}

func handle(e error) {
	if e != nil {
		panic(e)
	}
}
