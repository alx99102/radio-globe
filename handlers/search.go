package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

func HandleSearchInputChange(w http.ResponseWriter, r *http.Request) {
	// get the location from element's value
	if err := r.ParseForm(); err != nil {
		fmt.Println("Error parsing form:", err)
		return
	}

	var location string
	var temp string

	for i := 0; i < 5; i++ {
		temp = r.FormValue("location-choice-" + strconv.Itoa(i))
		if temp != "" {

			location = temp
			break
		}
	}

	// Parse searchInput.html template
	tmpl := template.Must(template.ParseFiles("templates/searchInput.html"))

	// Generate html fragment
	locationFrag := ""
	if location == "" {
		locationFrag = "No location entered"
	} else {
		locationFrag = location
	}

	err := tmpl.Execute(w, locationFrag)
	if err != nil {
		fmt.Println("Error executing template:", err)
	}
}