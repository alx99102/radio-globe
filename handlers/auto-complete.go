package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"radio-globe/credentials"
	"radio-globe/data"
)


func AutoCompleteSearch(w http.ResponseWriter, r *http.Request) {
	location := r.URL.Query().Get("location")
	if location == "" {
		// Handle empty input scenario
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
	var featureCollection data.FeatureCollection
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
	tmpl := template.Must(template.ParseFiles("templates/location.html"))
	err = tmpl.Execute(w, suggestionsMap)
	if err != nil {
		fmt.Println("Error executing template:", err)
	}
}

func callAutoComplete(location string) []byte {
	// Construct the URL
	baseUrl := "https://api.geoapify.com/v1/geocode/autocomplete"
	credsString, err := ioutil.ReadFile("credentials/creds.json")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var credentials credentials.Credentials
	err = json.Unmarshal(credsString, &credentials)
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
		fmt.Println(err)
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