package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

type FeatureCollection struct {
	Features []Feature `json:"features"`
}

type Feature struct {
	Properties Properties `json:"properties"`
}

type Properties struct {
	Name         string `json:"name"`
	Country      string `json:"country"`
	Region       string `json:"region"`
	State        string `json:"state"`
	City         string `json:"city"`
	Formatted    string `json:"formatted"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))

	tmpl.Execute(w, nil)
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	location := r.PostFormValue("location")

	w.Write([]byte(location))
}

func autoCompleteSearch(w http.ResponseWriter, r *http.Request) {
	location := r.URL.Query().Get("location")
	if location == "" {
		// Handle empty input scenario, perhaps by sending an empty response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
		return
	}
	autoCompletesJsonData := callAutoComplete(location)

	// Unmarshal JSON into your structs
	var featureCollection FeatureCollection
	err := json.Unmarshal([]byte(autoCompletesJsonData), &featureCollection)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	addressStrings := []string{}

	// Iterate through the features to extract the address
	for _, feature := range featureCollection.Features {
		addressStrings = append(addressStrings, feature.Properties.Formatted)
	}

	// Generate html fragment for each address
	html := ""
	for _, address := range addressStrings {
		html += "<div>" + address + "</div>"
	}

	// Send the html fragment as response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func callAutoComplete(location string) []byte {
	baseUrl := "https://api.geoapify.com/v1/geocode/autocomplete"
	apiKey := ""
	url := fmt.Sprintf("%s?text=%s&apiKey=%s", baseUrl, location, apiKey)

	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err); 
		return nil
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return body
}

func main() {
	fmt.Println("Starting server...")

	http.HandleFunc("/search/", handleSearch)
	http.HandleFunc("/auto-complete/", autoCompleteSearch)
	http.HandleFunc("/", handler)

	log.Fatal(http.ListenAndServe(":8080", nil))

	fmt.Println("Server started")
}
