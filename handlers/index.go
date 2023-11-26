package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"radio-globe/credentials"
)

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	creds, err := credentials.GetCredentials()
	if err != nil {
		fmt.Println("Error getting credentials:", err)
		return
	}
	
	// Generate html fragment and populate with credentials
	err = tmpl.Execute(w, creds)
	if err != nil {
		fmt.Println("Error executing template:", err)
	}
}