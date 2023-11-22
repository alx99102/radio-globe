package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

type Credentials struct {
	GeoapifyApiKey string `json:"geoapify_api_key"`
	GoogleMapsApiKey string `json:"google_maps_api_key"`
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

var db *sql.DB

func handler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))

	var credentials Credentials
	data, err := ioutil.ReadFile("creds.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(data, &credentials)
	if err != nil {
		fmt.Println(err)
		return
	}

	
	CredMap := map[string]string{
    	"APIKey": credentials.GoogleMapsApiKey,
	}
	// Generate html fragment and populate with credentials
	err = tmpl.Execute(w, CredMap)
	if err != nil {
		fmt.Println("Error executing template:", err)
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {

	// get the location from element's value
	if err := r.ParseForm(); err != nil {
		fmt.Println("Error parsing form:", err)
		return
	}

	// location := r.FormValue("location-button")
	temp := ""
	location := ""
	for i := 0; i < 5; i++ {
		// fmt.Print("location-choice-" + strconv.Itoa(i) + "\n")
		temp = r.FormValue("location-choice-" + strconv.Itoa(i))
		if temp != "" {

			location = temp
			break
		}
	}
	locationFrag := `<input 
	class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline mr-4" 
	type="text" 
	id="autocomplete" 
	name="location" 
	placeholder="Enter location" 
	hx-get="/auto-complete/" 
	hx-trigger="keyup changed delay:1000ms" 
	hx-target="#suggestions" 
	hx-indicator="#loading"
	value="%s">
	<button 
	class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline" 
	type="submit">
	Submit
	</button>
	`
	locationFrag = fmt.Sprintf(locationFrag, location)
	w.Write([]byte(locationFrag))
}

func autoCompleteSearch(w http.ResponseWriter, r *http.Request) {
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
	err = tmpl.Execute(w, suggestionsMap)
	if err != nil {
		fmt.Println("Error executing template:", err)
	}
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

func mapChangeFunc(w http.ResponseWriter, r *http.Request) {
	var credentials Credentials
	data, err := ioutil.ReadFile("creds.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(data, &credentials)
	if err != nil {
		fmt.Println(err)
		return
	}
	tmpl := `<iframe 
	class="w-screen h-screen"
	style="border:0"
	loading="lazy"
	allowfullscreen
	referrerpolicy="no-referrer-when-downgrade"
	src="https://www.google.com/maps/embed/v1/place?key=%s
		&q=%s">
	</iframe>`
	location := r.URL.Query().Get("location")
	if location == "" {
		location = "this is a gibberish string" 
	}
	tmpl = fmt.Sprintf(tmpl, credentials.GoogleMapsApiKey, location)
	w.Write([]byte(tmpl))
}

func main() {
	fmt.Println("Opening database")
	db = initDB()
	defer db.Close()

	fmt.Print("Ingesting database")

	// Read JSON data from file
	jsonData, err := ioutil.ReadFile("../out_filtered.json")
	if err != nil {
		log.Fatal(err)
	}
	

	var stations []Station
	err = json.Unmarshal(jsonData, &stations)
	if err != nil {
   		log.Fatal(err)
	}
	// ingestDB(db, stations)

	fmt.Println("Database created successfully")
	fmt.Println("Starting server")

	http.HandleFunc("/search/", handleSearch)
	http.HandleFunc("/auto-complete/", autoCompleteSearch)
	http.HandleFunc("/", handler)
	http.HandleFunc("/map/", mapChangeFunc)

	// listen for CTRL+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		for {
			sig := <-sigChan
			switch sig {
			case os.Interrupt:
				// Handle Ctrl+C: Close the database connection
				fmt.Println("Closing database connection and shutting down")
				db.Close()
				os.Exit(0)
			}
		}
	}()
	
	log.Fatal(http.ListenAndServe(":8080", nil))

}
