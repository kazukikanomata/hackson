package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbpool, err := InitDB(ctx)
	if err != nil {
		log.Fatalf("エラー: %v", err)
	}
	defer dbpool.Close()

	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {

		if err := dbpool.PingContext(r.Context()); err != nil {
			http.Error(w, `{"status": "db connection failed"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status": "ok"}`)
	})

	http.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"message": "Hello, Go API!!"}`)
	})

	fmt.Println("Go server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
