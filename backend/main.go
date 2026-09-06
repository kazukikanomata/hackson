package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status": "ok"}`)
	})

	http.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"message": "Hello, Go API!!"}`)
	})

	fmt.Println("Go server starting on :8080...")
	// エラーを握り潰すとポート競合時に無言で終了してしまうため必ず出力する
	log.Fatal(http.ListenAndServe(":8080", nil))
}
