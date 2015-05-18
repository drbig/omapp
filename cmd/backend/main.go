// See LICENSE.txt for licensing information.

package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	//_ "github.com/go-sql-driver/mysql"
	//_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"omapp/vendor/env"

	"omapp/pkg/http/logging"
	"omapp/pkg/http/reply"
	"omapp/pkg/model"
)

const (
	PAGE = 30
)

var (
	router *mux.Router
	store  sessions.Store
)

func main() {
	log.Println("Starting backend...")
	addr := env.StringDefault("OMA_WEB_MOUNT", "0.0.0.0:7777")
	store = sessions.NewCookieStore([]byte(env.String("OMA_WEB_SECRET")))
	log.Println("Connecting to database...")
	if err := model.Init(); err != nil {
		log.Fatalln(err)
		os.Exit(3)
	}
	router = mux.NewRouter()
	router.HandleFunc("/user/{login}", handleUserCheck).Methods("GET")
	router.HandleFunc("/user/{login}", handleUserAuth).Methods("POST")
	router.HandleFunc("/user/{login}/info", handleUserInfo)
	router.HandleFunc("/browse/{by}", handleBrowse)
	router.HandleFunc("/info", handleInfo)
	log.Println("Firing up HTTP server at", addr)
	log.Fatalln(http.ListenAndServe(addr,
		context.ClearHandler(logging.Handler(router)),
	))
}

func handleUserCheck(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	login := v["login"]
	var user model.User
	if err := model.Db.Where("login = ?", login).First(&user).Error; err != nil {
		reply.Send(w, http.StatusOK, true, false)
		return
	}
	reply.Send(w, http.StatusOK, true, true)
}

func handleUserAuth(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println("ERROR:", err)
		reply.Send(w, http.StatusInternalServerError, false, "error parsing form")
		return
	}
	pass := r.PostForm.Get("password")
	if pass == "" {
		reply.Send(w, http.StatusOK, false, "no password given")
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
			reply.Send(w, http.StatusInternalServerError, false, "error creating user")
			return
		}
		userSetCookie(w, r, user.ID)
		reply.Send(w, http.StatusOK, true, "user added")
		return
	}
	if !user.CheckPassword(pass) {
		reply.Send(w, http.StatusOK, false, "wrong password")
		return
	}
	userSetCookie(w, r, user.ID)
	reply.Send(w, http.StatusOK, true, "auth ok")
}

func userSetCookie(w http.ResponseWriter, r *http.Request, userid int) {
	s, _ := store.Get(r, "omapp-session")
	s.Values["user"] = userid
	s.Save(r, w)
}

func handleUserInfo(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	login := v["login"]
	var user model.User
	if err := model.Db.Where("login = ?", login).First(&user).Error; err != nil {
		reply.Send(w, http.StatusOK, false, "no such user")
		return
	}
	var maps []model.Map
	if err := model.Db.Model(&user).Related(&maps).Order("state desc, created_at desc").Error; err != nil {
		log.Println("ERROR:", err)
		reply.Send(w, http.StatusInternalServerError, false, "error fetching user maps")
		return
	}
	reply.Send(w, http.StatusOK, true, map[string]interface{}{
		"user": user,
		"maps": maps,
	})
}

func handleBrowse(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println("ERROR:", err)
		reply.Send(w, http.StatusInternalServerError, false, "error parsing form")
		return
	}
	page, err := strconv.Atoi(r.Form.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	v := mux.Vars(r)
	by := v["by"]
	query := model.Db.Table("maps").Where("state = ?", model.READY)
	switch by {
	case "date":
		query.Order("created_at desc")
	case "area":
		query.Order("area desc, created_at desc")
	case "visited":
		query.Order("visited desc, created_at desc")
	default:
		log.Println("WARN: no such browse order")
		reply.Send(w, http.StatusOK, false, "no such browse order")
		return
	}
	var maps []model.MapPublic
	var total int
	query.Count(&total).Offset((page - 1) * PAGE).Limit(PAGE)
	query.Select("users.login, maps.*")
	query.Joins("left join users on users.id = maps.user_id").Scan(&maps)
	if query.Error != nil {
		log.Println("ERROR:", query.Error)
		reply.Send(w, http.StatusInternalServerError, false, "error running query")
		return
	}
	reply.Send(w, http.StatusOK, true, map[string]interface{}{
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
		reply.Send(w, http.StatusInternalServerError, false, "error running query")
		return
	}
	var queue []model.MapPublic
	var qtotal int
	query = model.Db.Table("maps").Where("state = ?", model.QUEUE).Order("created_at desc")
	query.Count(&qtotal).Limit(10).Select("users.login, maps.*")
	query.Joins("left join users on users.id = maps.user_id").Scan(&queue)
	if query.Error != nil {
		log.Println("ERROR:", query.Error)
		reply.Send(w, http.StatusInternalServerError, false, "error running query")
		return
	}
	reply.Send(w, http.StatusOK, true, map[string]interface{}{
		"maps": maps, "mtotal": mtotal,
		"queue": queue, "qtotal": qtotal,
	})
}
