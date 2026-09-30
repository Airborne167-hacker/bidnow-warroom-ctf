package main

import (
	"log"
	"net/http"
)

func main() {
	initDB()
	loadTemplates()

	mux := http.NewServeMux()

	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/account", accountHandler)
	mux.HandleFunc("/paddle-status", paddleStatusHandler)
	mux.HandleFunc("/lot/", lotHandler)

	// deliberately exposed / vulnerable surfaces
	mux.HandleFunc("/sitemap.xml", sitemapHandler)
	mux.HandleFunc("/backups/paddle-export-2026.csv", backupFileHandler)

	// War Room (the CTF layer)
	mux.HandleFunc("/warroom", warroomHandler)
	mux.HandleFunc("/api/hint", hintHandler)
	mux.HandleFunc("/api/submit", submitHandler)
	mux.HandleFunc("/api/progress", progressHandler)

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	addr := ":8083"
	log.Printf("BidNow Live War Room running: http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
