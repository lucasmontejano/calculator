// Command server starts the calculator REST API.
package main

import (
	"log"
	"net/http"
	"os"

	"calculator-app/backend/internal/httpapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("API listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, httpapi.NewRouter()))
}
