package qiao

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHttpStress(t *testing.T) {

	// r := Http.NewRouter()
	// r.Get("/version", GetVersion)
	// r.Get("/", home)
	http.HandleFunc("/version", Version)
	http.ListenAndServe(":8080", nil)
}

func Version(w http.ResponseWriter, r *http.Request) {
	var ver struct {
		Version string `json:"version"`
	}
	ver.Version = "1.0.0"
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	json.NewEncoder(w).Encode(ver)
}
