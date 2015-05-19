// See LICENSE.txt for licensing information.

package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/darkhelmet/env"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"omapp/pkg/model"
	"omapp/pkg/queue"
	"omapp/pkg/web"
)

var (
	root   string
	router *mux.Router
	store  sessions.Store
)

func main() {
	log.Println("Starting uploader...")
	root = env.String("OMA_DATA_ROOT")
	store = sessions.NewCookieStore([]byte(env.String("OMA_WEB_SECRET")))
	addr := env.StringDefault("OMA_U_MOUNT", "0.0.0.0:8777")
	log.Println("Connecting to database...")
	if err := model.Init(); err != nil {
		log.Fatalln(err)
		os.Exit(3)
	}
	log.Println("Connecting to queue...")
	if err := queue.Init(); err != nil {
		log.Fatalln(err)
		os.Exit(3)
	}
	router = mux.NewRouter()
	router.HandleFunc("/upload", handleUpload).Methods("POST")
	log.Println("Firing up HTTP server at", addr)
	log.Fatalln(http.ListenAndServe(addr,
		context.ClearHandler(web.Logging(router)),
	))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	s, _ := store.Get(r, "omapp-session")
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
	log.Println("Upload request from:", user.Login)
	mrec := model.Map{UserID: uid}
	if err := model.Db.Save(&mrec).Error; err != nil {
		log.Println("ERROR:", err)
		web.Reply(w, http.StatusInternalServerError, false, "database error")
		return
	}
	dir := path.Join(root, "uploads", strconv.Itoa(mrec.ID))
	if err := os.Mkdir(dir, 0777); err != nil {
		errorCleanup(w, &mrec, err, "filesystem error")
		return
	}
	var uploaded, skipped int
	mpr, err := r.MultipartReader()
	if err != nil {
		errorCleanup(w, &mrec, err, "multipart error")
		return
	}
	for {
		p, err := mpr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			errorCleanup(w, &mrec, err, "multipart decoder error")
			return
		}
		if p.FormName() == "worldname" {
			buf, err := ioutil.ReadAll(p)
			if err != nil {
				log.Println("WARN: reading worldname:", err)
			}
			wname := string(buf)
			log.Println(user.Login, "world name:", wname)
			mrec.WorldName = wname
			if err := model.Db.Save(&mrec).Error; err != nil {
				log.Println("ERROR:", err)
			}
			continue
		}
		name := p.FileName()
		if name == "" {
			skipped += 1
			log.Println("WARN: empty file name, skipping")
			continue
		}
		log.Println(user.Login, "uploading", name)
		target := path.Join(dir, name)
		fh, err := os.Create(target)
		if err != nil {
			errorCleanup(w, &mrec, err, "filesystem error")
			return
		}
		if _, err := io.Copy(fh, p); err != nil {
			fh.Close()
			errorCleanup(w, &mrec, err, "filesystem error")
			return
		}
		fh.Close()
		uploaded += 1
	}
	log.Println(user.Login, "uploaded:", uploaded, "skipped:", skipped)
	if _, err := queue.Send(queue.MAP, mrec.ID, 5, 5*time.Second); err != nil {
		log.Println("FATAL:", err)
	}
	web.Reply(w, http.StatusOK, true, map[string]int{
		"uploaded": uploaded,
		"skipped":  skipped,
	})
}

func errorCleanup(w http.ResponseWriter, mrec *model.Map, err error, msg string) {
	log.Println("ERROR:", err)
	mrec.State = model.BROKEN
	if err := model.Db.Save(mrec).Error; err != nil {
		log.Println("FATAL:", err)
	}
	if _, err := queue.Send(queue.CLEAN, mrec.ID, 10, 4*time.Hour); err != nil {
		log.Println("FATAL:", err)
	}
	web.Reply(w, http.StatusInternalServerError, false, msg)
}
