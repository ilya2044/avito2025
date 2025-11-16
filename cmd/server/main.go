package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/ilya2044/avito2025/internal/api"
	"github.com/ilya2044/avito2025/internal/service"
	"github.com/ilya2044/avito2025/internal/storage"
)

var migrationSQL string

func main() {
	flag.Parse()
	db, err := storage.NewDBFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	repo := storage.NewRepository(db)
	if err := repo.ApplyMigrations(migrationSQL); err != nil {
		log.Fatal("migration failed:", err)
	}
	svc := service.NewService(repo)
	h := api.NewHandler(svc)
	r := mux.NewRouter()
	h.RegisterRoutes(r)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	fmt.Println("listening on", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
