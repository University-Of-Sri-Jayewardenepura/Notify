package main

import (
	"log"
	"net/http"

	"github.com/pruthivithejan/notify/internal/config"
	"github.com/pruthivithejan/notify/internal/httpapi"
)

func main() {
	cfg := config.Load()

	log.Printf("starting notify on :%s", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, httpapi.NewRouter()); err != nil {
		log.Fatal(err)
	}
}
