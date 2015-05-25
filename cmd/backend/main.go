// See LICENSE.txt for licensing information.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/darkhelmet/env"
	"github.com/gorilla/mux"

	"omapp/pkg/model"
	"omapp/pkg/ver"
	"omapp/pkg/web"
)

const (
	PAGE = 30
)

func main() {
	log.Println("Starting backend version:", ver.VERSION)
	addr := env.StringDefault("OMA_B_MOUNT", "0.0.0.0:7777")
	log.Println("Connecting to database...")
	if err := model.Init(); err != nil {
		log.Fatalln(err)
		os.Exit(3)
	}
	web.Router.HandleFunc("/user", handleUser)
	web.Router.HandleFunc("/user/{login}", handleUserCheck).Methods("GET")
	web.Router.HandleFunc("/user/{login}", handleUserAuth).Methods("POST")
	web.Router.HandleFunc("/user/{login}/info", handleUserInfo)
	web.Router.HandleFunc("/browse/{by}", handleBrowse)
	web.Router.HandleFunc("/info", handleInfo)
	web.Router.HandleFunc("/map/{id}", handleMap)
	web.Fire(addr)
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	s, _ := web.Store.Get(r, "omapp-session")
	uidval, present := s.Values["user"]
	if !present {
		web.Reply(w, http.StatusOK, false, "not logged in")
		return
	}
	uid := uidval.(int)
	var user model.User
	if err := model.Db.First(&user, uid).Error; err != nil {
		web.Reply(w, http.StatusOK, false, "no such user")
		return
	}
	web.Reply(w, http.StatusOK, true, user.Login)
}

func handleUserCheck(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	login := v["login"]
	var user model.User
	if err := model.Db.Where("login = ?", login).First(&user).Error; err != nil {
		web.Reply(w, http.StatusOK, true, false)
		return
	}
	web.Reply(w, http.StatusOK, true, true)
}

func handleUserAuth(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println("ERROR:", err)
		web.Reply(w, http.StatusInternalServerError, false, "error parsing form")
		return
	}
	pass := r.PostForm.Get("password")
	if pass == "" {
		web.Reply(w, http.StatusOK, false, "no password given")
		return
	}
	v := mux.Vars(r)
	login := v["login"]
	var user model.User
	if err := model.Db.Where("login = ?", login).First(&user).Error; err != nil {
		user = model.User{Login: login}
		user.SetPassword(pass)
		if err := model.Db.Create(&user).Error; err != nil {
			log.Println("ERROR:", err)
			web.Reply(w, http.StatusInternalServerError, false, "error creating user")
			return
		}
		userSetCookie(w, r, user.ID)
		web.Reply(w, http.StatusOK, true, "user added")
		return
	}
	if !user.CheckPassword(pass) {
		web.Reply(w, http.StatusOK, false, "wrong password")
		return
	}
	userSetCookie(w, r, user.ID)
	web.Reply(w, http.StatusOK, true, "auth ok")
}

func userSetCookie(w http.ResponseWriter, r *http.Request, userid int) {
	s, _ := web.Store.Get(r, "omapp-session")
	s.Values["user"] = userid
	s.Save(r, w)
}

func handleUserInfo(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	login := v["login"]
	var user model.User
	if err := model.Db.Where("login = ?", login).First(&user).Error; err != nil {
		web.Reply(w, http.StatusOK, false, "no such user")
		return
	}
	var maps []model.Map
	if err := model.Db.Model(&user).Related(&maps).Order("state desc, created_at desc").Error; err != nil {
		log.Println("ERROR:", err)
		web.Reply(w, http.StatusInternalServerError, false, "error fetching user maps")
		return
	}
	web.Reply(w, http.StatusOK, true, map[string]interface{}{
		"user": user,
		"maps": maps,
	})
}

func handleBrowse(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println("ERROR:", err)
		web.Reply(w, http.StatusInternalServerError, false, "error parsing form")
		return
	}
	page, err := strconv.Atoi(r.Form.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	v := mux.Vars(r)
	by := v["by"]
	var order string
	switch by {
	case "date":
		order = "created_at desc"
	case "area":
		order = "area desc, created_at desc"
	case "visited":
		order = "visited desc, created_at desc"
	default:
		log.Println("WARN: no such browse order")
		order = "created_at desc"
	}
	offset := (page - 1) * PAGE
	query := model.Db.Raw(
		fmt.Sprintf(
			"select users.login, maps.* from maps left join users on users.id = maps.user_id where state = %d order by %s limit %d offset %d",
			model.READY,
			order,
			PAGE,
			offset,
		),
	)
	var maps []model.MapPublic
	query.Scan(&maps)
	if query.Error != nil {
		log.Println("ERROR:", query.Error)
		web.Reply(w, http.StatusInternalServerError, false, "error running query")
		return
	}
	var total int
	model.Db.Table("maps").Where("state = ?", model.READY).Count(&total)
	web.Reply(w, http.StatusOK, true, map[string]interface{}{
		"maps":  maps,
		"page":  page,
		"pages": total/PAGE + 1,
	})
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	var maps []model.MapPublic
	var mtotal int

	query := model.Db.Table("maps").Where("state = ?", model.READY).Order("created_at desc")
	query.Count(&mtotal).Limit(10).Select("users.login, maps.*")
	query.Joins("left join users on users.id = maps.user_id").Scan(&maps)
	if query.Error != nil {
		log.Println("ERROR:", query.Error)
		web.Reply(w, http.StatusInternalServerError, false, "error running query")
		return
	}
	var queue []model.MapPublic
	var qtotal int
	query = model.Db.Table("maps").Where("state = ?", model.QUEUE).Order("created_at desc")
	query.Count(&qtotal).Limit(10).Select("users.login, maps.*")
	query.Joins("left join users on users.id = maps.user_id").Scan(&queue)
	if query.Error != nil {
		log.Println("ERROR:", query.Error)
		web.Reply(w, http.StatusInternalServerError, false, "error running query")
		return
	}
	web.Reply(w, http.StatusOK, true, map[string]interface{}{
		"maps": maps, "mtotal": mtotal,
		"queue": queue, "qtotal": qtotal,
	})
}

func handleMap(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]
	query := model.Db.Raw(
		fmt.Sprintf(
			"select users.login, maps.* from maps left join users on users.id = maps.user_id where maps.id = %s",
			id,
		),
	)
	var maps []model.MapPublic
	query.Scan(&maps)
	if query.Error != nil {
		log.Println("ERROR:", query.Error)
		web.Reply(w, http.StatusInternalServerError, false, "error running query")
		return
	}
	if len(maps) < 1 {
		web.Reply(w, http.StatusOK, false, "map not found")
		return
	}
	web.Reply(w, http.StatusOK, true, maps[0])
}
