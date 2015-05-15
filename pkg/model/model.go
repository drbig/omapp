// See LICENSE.txt for licensing information.

package model

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"time"

	"github.com/jinzhu/gorm"

	"omapp/vendor/env"
)

const (
	VERSION = "3"
)

type User struct {
	ID        int
	Login     string `sql:"unique_index"`
	Password  string
	Salt      uint32
	CreatedAt time.Time
	RetiredAt *time.Time
	Maps      []Map
}

func (u *User) BeforeCreate() error {
	u.CreatedAt = time.Now()
	return nil
}

func (u *User) SetPassword(plain string) {
	u.Salt = rand.Uint32()
	u.Password = getPassHash(plain, u.Salt)
}

func (u *User) CheckPassword(plain string) bool {
	hash := getPassHash(plain, u.Salt)
	if hash == u.Password {
		return true
	}
	return false
}

type MapState int

const (
	QUEUE MapState = iota
	READY
	BROKEN
	DELETE
)

type Map struct {
	ID            int
	UserID        int `sql:"index"`
	State         MapState
	CreatedAt     time.Time
	ImageName     string
	WorldName     string
	Width, Height int
	Area, Visited int
}

func (m *Map) BeforeCreate() error {
	m.State = QUEUE
	m.CreatedAt = time.Now()
	return nil
}

var Db gorm.DB

func Init() error {
	var err error
	Db, err = gorm.Open(env.String("OMA_DB_DRIVER"), env.String("OMA_DB_ARGS"))
	if err != nil {
		return err
	}
	if env.BoolDefault("OMA_DB_VERBOSE", false) {
		fmt.Println("DB VERBOSE")
		Db.LogMode(true)
	}
	return nil
}

func getPassHash(plain string, salt uint32) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(fmt.Sprintf("%d%s", salt, plain))))
}
