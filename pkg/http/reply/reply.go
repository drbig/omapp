// See LICENSE.txt for licensing information.

package reply

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

func Send(w http.ResponseWriter, status int, success bool, data interface{}) {
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
