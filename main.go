package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"radio-globe/data"
	"radio-globe/handlers"
)

func main() {
	fmt.Println("Opening database")
	data.InitDB()
	

	fmt.Println("Database opened successfully")
	fmt.Println("Starting server")

	http.HandleFunc("/", handlers.HandleIndex)
	http.HandleFunc("/search/", handlers.HandleSearchInputChange)
	http.HandleFunc("/auto-complete/", handlers.AutoCompleteSearch)
	http.HandleFunc("/map/", handlers.MainContentChangeHandler)
	http.HandleFunc("/static/favicon.png", handlers.ServeFavicon)
	http.HandleFunc("/static/spinner.svg", handlers.ServeSpinner)

	// listen for CTRL+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		for {
			sig := <-sigChan
			if(sig != nil) {
				fmt.Println("Closing database connection and shutting down")
				data.CloseDB()
				os.Exit(0)
			}
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", nil))
}