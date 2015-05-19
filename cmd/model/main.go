// See LICENSE.txt for licensing information.

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jwaldrip/odin/cli"

	"omapp/pkg/model"
	"omapp/pkg/ver"
)

var cmd = cli.New(ver.VERSION, "Model utility", check)

func init() {
	cmd.DefineSubCommand("migrate", "auto migrate database", migrate)
	scmd := cmd.DefineSubCommand("destroy", "destroy database", destroy)
	scmd.DefineBoolFlag("confirm", false, "please confirm")
	cmd.DefineSubCommand("adduser", "add an user", adduser, "login", "password")
	cmd.DefineSubCommand("retire", "retire an user", retire, "login")
}

func main() {
	cmd.Start()
}

func check(c cli.Command) {
	fmt.Println("ENV variables:")
	fmt.Println("  OMA_DB_DRIVER:", os.Getenv("OMA_DB_DRIVER"))
	fmt.Println("    OMA_DB_ARGS:", os.Getenv("OMA_DB_ARGS"))
	fmt.Println(" OMA_DB_VERBOSE:", os.Getenv("OMA_DB_VERBOSE"))
}

func migrate(c cli.Command) {
	fmt.Println("Attempting auto-migrate...")
	connect()
	model.Db.AutoMigrate(&model.User{}, &model.Map{})
	disconnect()
}

func destroy(c cli.Command) {
	fmt.Println("Attempting destroy all tables...")
	if c.Flag("confirm").Get() != true {
		fmt.Println("Not confirmed, bailing.")
		os.Exit(1)
	}
	connect()
	model.Db.DropTable(&model.Map{})
	model.Db.DropTable(&model.User{})
	disconnect()
}

func adduser(c cli.Command) {
	login := c.Param("login").String()
	pass := c.Param("password").String()
	fmt.Println("Attempting to add user...")
	connect()
	user := model.User{Login: login}
	user.SetPassword(pass)
	model.Db.Create(&user)
	disconnect()
}

func retire(c cli.Command) {
	var user model.User
	login := c.Param("login").String()
	fmt.Println("Attempting to retire user...")
	connect()
	model.Db.Where("login = ?", login).First(&user)
	if user.RetiredAt != nil {
		fmt.Println("FAILED user", login, "already retired.")
	} else {
		now := time.Now()
		user.RetiredAt = &now
		model.Db.Save(&user)
	}
	disconnect()
}

func connect() {
	fmt.Println("Connecting to database...")
	if err := model.Init(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	model.Db.LogMode(true)
}

func disconnect() {
	fmt.Println("Closing database...")
	model.Db.Close()
}
