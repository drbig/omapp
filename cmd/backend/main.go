// See LICENSE.txt for licensing information.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/darkhelmet/env"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	//_ "github.com/go-sql-driver/mysql"
	//_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"omapp/pkg/http/logging"
	"omapp/pkg/model"
)

const (
	VERSION = "1"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

var (
	router *mux.Router
	store  sessions.Store
)

func main() {
	log.Println("Starting backend...")

	store = sessions.NewCookieStore([]byte(env.String("OMA_WEB_SECRET")))

	if err := model.Init(); err != nil {
		log.Fatalln(err)
		os.Exit(3)
	}

	router = mux.NewRouter()
	router.HandleFunc("/user/{login}", handleUserCheck).Methods("GET")
	router.HandleFunc("/user/{login}", handleUserAuth).Methods("POST")
	router.HandleFunc("/user/{login}/info", handleUserInfo)

	addr := fmt.Sprintf("%s:%s",
		env.StringDefault("OMA_WEB_HOST", "0.0.0.0"),
		env.StringDefault("OMA_WEB_PORT", "7777"),
	)
	log.Println("Firing up HTTP server at", addr)
	log.Fatalln(http.ListenAndServe(addr,
		context.ClearHandler(logging.Handler(router)),
	))
}

func reply(w http.ResponseWriter, status int, success bool, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	raw, err := json.Marshal(Response{success, data})
	if err != nil {
		log.Println("ERROR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "data": "error generating reply"}`))
		return
	}
	w.WriteHeader(status)
	w.Write(raw)
}

func handleUserCheck(w http.ResponseWriter, r *http.Request) {
	var user model.User
	v := mux.Vars(r)
	login := v["login"]
	if err := model.Db.Where("login = ?", login).First(&user).Error; err != nil {
		reply(w, http.StatusOK, true, false)
		return
	}
	reply(w, http.StatusOK, true, true)
}

func handleUserAuth(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := r.ParseForm(); err != nil {
		log.Println("ERROR:", err)
		reply(w, http.StatusInternalServerError, false, "error parsing form")
		return
	}
	pass := r.PostForm.Get("password")
	if pass == "" {
		reply(w, http.StatusOK, false, "no password given")
		return
	}
	v := mux.Vars(r)
	login := v["login"]

	if err := model.Db.Where("login = ?", login).First(&user).Error; err != nil {
		user = model.User{Login: login}
		user.SetPassword(pass)
		if err := model.Db.Create(&user).Error; err != nil {
			log.Println("ERROR:", err)
			reply(w, http.StatusInternalServerError, false, "error creating user")
			return
		}
		userSetCookie(w, r, user.ID)
		reply(w, http.StatusOK, true, "user added")
		return
	}

	if !user.CheckPassword(pass) {
		reply(w, http.StatusOK, false, "wrong password")
		return
	}
	userSetCookie(w, r, user.ID)
	reply(w, http.StatusOK, true, "auth ok")
}

func userSetCookie(w http.ResponseWriter, r *http.Request, userid int) {
	s, _ := store.Get(r, "omapp-session")
	s.Values["user"] = userid
	s.Save(r, w)
}

func handleUserInfo(w http.ResponseWriter, r *http.Request) {
	var user model.User
	var maps []model.Map
	v := mux.Vars(r)
	login := v["login"]
	if err := model.Db.Where("login = ?", login).First(&user).Error; err != nil {
		reply(w, http.StatusOK, false, "no such user")
		return
	}
	if err := model.Db.Model(&user).Related(&maps).Error; err != nil {
		log.Println("ERROR:", err)
		reply(w, http.StatusInternalServerError, false, "error fetching user maps")
		return
	}
	reply(w, http.StatusOK, true, map[string]interface{}{
		"user": map[string]interface{}{
			"login":   user.Login,
			"since":   user.CreatedAt,
			"retired": user.RetiredAt,
		},
		"maps": maps,
	})
}
