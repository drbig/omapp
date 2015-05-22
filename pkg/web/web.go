// See LICENSE.txt for licensing information.

package web

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/darkhelmet/env"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"omapp/pkg/ver"
)

var (
	Store  = sessions.NewCookieStore([]byte(env.String("OMA_WEB_SECRET")))
	Router = mux.NewRouter()
)

func Fire(addr string) {
	log.Println("Firing up HTTP server at", addr)
	log.Fatalln(http.ListenAndServe(addr,
		context.ClearHandler(Logging(Router)),
	))
}

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	w.length = len(b)
	return w.ResponseWriter.Write(b)
}

func Logging(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := statusWriter{w, 0, 0}
		h.ServeHTTP(&sw, r)
		end := time.Now()
		duration := end.Sub(start)
		log.Println(r.RemoteAddr, r.Method, r.URL, sw.status, sw.length, duration)
	}
}

type Message struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Version string      `json:"ver"`
}

func Reply(w http.ResponseWriter, status int, success bool, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:9191") // VERY TEMPORARY
	w.Header().Set("Access-Control-Allow-Credentials", "true")             // VERY TEMPORARY
	raw, err := json.Marshal(Message{success, data, ver.VERSION})
	if err != nil {
		log.Println("ERROR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "data": "error generating reply"}`))
		return
	}
	w.WriteHeader(status)
	w.Write(raw)
}
