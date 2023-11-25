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
	GeoapifyApiKey   string `json:"geoapify_api_key"`
	GoogleMapsApiKey string `json:"google_maps_api_key"`
	HereApiKey       string `json:"here_api_key"`
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

	stationsArr := []Station{}
	location := r.URL.Query().Get("location")
	if location == "" {
		location = "this is a gibberish string"
	} else {
		lat, lon, err := geocode(location)
		if err != nil {
			fmt.Println(err)
			return
		}
		stationsArr = searchByCoordinates(db, lat, lon)
	}

	//stations := `<div class="max-h-screen overflow-y-auto">`
	var stations string
	for _, station := range stationsArr {
		contentType, err := getContentType(station.URLResolved)
		if err != nil {
			fmt.Println(err)
			continue
		}

		stations += fmt.Sprintf(
			`	<div class="bg-neutral-50 text-black py-2 px-4 focus:outline-none focus:shadow-outline text-center flex flex-row gap-4 border-b border-gray-300">
				<img src="%s" alt="station logo" class="w-1/4 h-1/4 rounded-lg">
				<div class="w-3/4 h-1/4 flex flex-col justify-center break-words">
				<div class="flex flex-col justify-center items-center">
					<span class="font-bold">
					%s<br>
					</span>
					<span class="text-gray-600">
					Language: %s
					</span>
					<audio controls class="w-4/5">
						<source src="%s" type="%s">
						Unable to display audio
					</audio>
				</div>
				</div>
			</div>`, station.Favicon, station.Name, station.Language, station.URLResolved, contentType)
	}
	//stations += `</div>`

	tmpl := `
	<div class="absolute top-0 right-0 w-1/4 max-h-screen overflow-y-auto">
		<div class="bg-zinc-200 text-black font-bold py-2 px-4 focus:outline-none focus:shadow-outline text-center">
			Radio List
		</div>
		<div class="list">
			%s
		</div>
	</div>
	<iframe 
		class="w-screen h-screen"
		style="border:0"
		loading="lazy"
		allowfullscreen
		referrerpolicy="no-referrer-when-downgrade"
		src="https://www.google.com/maps/embed/v1/place?key=%s&q=%s">
	</iframe>
	`
	tmpl = fmt.Sprintf(tmpl, stations, credentials.GoogleMapsApiKey, location)
	w.Write([]byte(tmpl))
}

func main() {
	fmt.Println("Opening database")
	db = initDB()
	defer db.Close()

	fmt.Println("Database opened successfully")
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

				// close all transactions
				rows, err := db.Query("SELECT id FROM Radios")
				if err != nil {
					log.Fatal(err)
				}

				rows.Close()
				db.Close()

				os.Exit(0)
			}
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", nil))

}
