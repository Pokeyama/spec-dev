package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	store := NewNoteStore(50, time.Now)
	server := NewServer(store)

	addr := ":8080"
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatal(err)
	}
}
