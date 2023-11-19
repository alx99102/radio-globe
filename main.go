package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

type Credentials struct {
	GeoapifyApiKey string `json:"geoapify_api_key"`
}

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


	// Call the autocomplete API
	autoCompletesJsonData := callAutoComplete(location)
	if autoCompletesJsonData == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error calling autocomplete API"))
		return
	}



	// Unmarshal JSON into struct
	var featureCollection FeatureCollection
	err := json.Unmarshal([]byte(autoCompletesJsonData), &featureCollection)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}


	// Format the address
	var formattedAddresses []string
	for _, feature := range featureCollection.Features {
    	formattedAddresses = append(formattedAddresses, feature.Properties.Formatted)
	}

	suggestionsMap := map[string][]string{
    	"Addresses": formattedAddresses,
	}

	// Generate html fragment 
	tmpl := template.Must(template.ParseFiles("location.html"))
	tmpl.Execute(w, suggestionsMap)
	
}

func callAutoComplete(location string) []byte {
	// Construct the URL
	baseUrl := "https://api.geoapify.com/v1/geocode/autocomplete"
	data, err := ioutil.ReadFile("creds.json")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var credentials Credentials
	err = json.Unmarshal(data, &credentials)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	apiKey := credentials.GeoapifyApiKey 
	url := fmt.Sprintf("%s?text=%s&apiKey=%s", baseUrl, location, apiKey)

	// Call the API
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
	if res.StatusCode != http.StatusOK {
        fmt.Printf("API request failed with status code: %d\n", res.StatusCode)
        return nil
    }
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return body
}

func main() {
	fmt.Println("Starting server")

	http.HandleFunc("/search/", handleSearch)
	http.HandleFunc("/auto-complete/", autoCompleteSearch)
	http.HandleFunc("/", handler)

	log.Fatal(http.ListenAndServe(":8080", nil))

}
