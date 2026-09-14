package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"ashen-crown-dungeon/internal/game"
	"ashen-crown-dungeon/internal/httpapi"
)

func main() {
	port := flag.Int("port", 8080, "http port")
	data := flag.String("data", "./data", "data directory")
	web := flag.String("web", "./web", "web directory")
	flag.Parse()
	if v := os.Getenv("PORT"); v != "" {
		*port = httpapi.ParsePort(v)
	}
	absData, _ := filepath.Abs(*data)
	absWeb, _ := filepath.Abs(*web)
	store := game.NewStore(absData)
	if overrides, err := store.LoadContentOverrides(); err == nil {
		for _, ov := range overrides {
			if err := game.ApplyContentOverride(ov); err != nil {
				log.Printf("skip editor override %s/%s: %v", ov.Kind, ov.ID, err)
			}
		}
	} else {
		log.Printf("load editor overrides: %v", err)
	}
	engine := game.NewEngine()
	server := httpapi.New(engine, store, absWeb)
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Ashen Crown running at http://localhost:%d", *port)
	log.Printf("data: %s", absData)
	if err := http.ListenAndServe(addr, server.Handler()); err != nil {
		log.Fatal(err)
	}
}
