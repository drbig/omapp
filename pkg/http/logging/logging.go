// See LICENSE.txt for licensing information.

package logging

import (
	"log"
	"net/http"
	"time"
)

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

func Handler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := statusWriter{w, 0, 0}
		h.ServeHTTP(&sw, r)
		end := time.Now()
		duration := end.Sub(start)
		log.Println(r.RemoteAddr, r.Method, r.URL, sw.status, sw.length, duration)
	}
}
