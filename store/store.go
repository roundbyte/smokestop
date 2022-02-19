package store

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"math/rand"
	"os"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	guuid "github.com/google/uuid"
	mailer "github.com/roundbyte/smokestop/mailer"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	EmailAddr      string
	Username       string
	Password       []byte
	Active         bool
	ActivationCode string
	StoppedAt      time.Time
}

type Store struct {
	Users map[string]User
}

func New() *Store {
	store := &Store{}
	store.Users = make(map[string]User)
	return store
}

func dbConnect() (*badger.DB, error) {
	opts := badger.DefaultOptions(os.Getenv("DBPATH"))
	opts.Logger = nil
	return badger.Open(opts)
}

type UserRegistrationForm struct {
	EmailAddr string `json:"emailAddr"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

func (store *Store) RegisterUser(userRegistrationForm UserRegistrationForm) (string, error) {
	db, err := dbConnect()
	if err != nil {
		return "", err
	}
	defer db.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userRegistrationForm.Password), 8)
	if err != nil {
		return "", err
	}
	userEncode := User{
		EmailAddr:      userRegistrationForm.EmailAddr,
		Username:       userRegistrationForm.Username,
		Password:       hashedPassword,
		Active:         false,
		ActivationCode: randSeq(10),
		StoppedAt:      time.Now(),
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

	mailer.SendEmail(userEncode.Username, userEncode.EmailAddr, userEncode.ActivationCode)
	return id, nil
}

func (store *Store) CheckPassword(userId string, password string) error {
	db, err := dbConnect()
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
			return errors.New("errPasswordMismatch")
		}
		return nil
	})
}

func (store *Store) VerifyUser(userId string, code string) error {
	db, err := dbConnect()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(txn *badger.Txn) error {
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
		if match := (decodedUser.ActivationCode == code); match == false {
			return errors.New("errInvalidCode")
		}
		decodedUser.Active = true
		decodedUser.ActivationCode = "isAlreadyActive"
		var byteEncodedUser bytes.Buffer
		encoder := gob.NewEncoder(&byteEncodedUser)
		if err := encoder.Encode(decodedUser); err != nil {
			return err
		}
		if err := txn.Set([]byte(userId), byteEncodedUser.Bytes()); err != nil {
			return err
		}
		return nil
	})
}

func (store *Store) GetAllUsers() error {
	db, err := dbConnect()
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

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
