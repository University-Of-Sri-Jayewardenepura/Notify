package main

import (
	"log"
	"net/http"

	"github.com/University-Of-Sri-Jayewardenepura/Notify/internal/config"
	"github.com/University-Of-Sri-Jayewardenepura/Notify/internal/httpapi"
	"github.com/University-Of-Sri-Jayewardenepura/Notify/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("starting notify on :%s", cfg.Port)

	svc := service.New(cfg.GitHubOrganization, nil)

	if err := http.ListenAndServe(":"+cfg.Port, httpapi.NewRouter(httpapi.RouterDependencies{
		GitHubWebhookSecret: cfg.GitHubWebhookSecret,
		GitHubService:       svc,
	})); err != nil {
		log.Fatal(err)
	}
}
