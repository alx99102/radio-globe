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

func MainContentChangeHandler(w http.ResponseWriter, r *http.Request) {
	var credentials credentials.Credentials
	credsString, err := ioutil.ReadFile("credentials/creds.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(credsString, &credentials)
	if err != nil {
		fmt.Println(err)
		return
	}

	stationsArr := []data.Station{}
	location := r.URL.Query().Get("location")
	if location == "" {
		location = "this is a gibberish string"
	} else {
		lat, lon, err := data.Geocode(location)
		if err != nil {
			fmt.Println(err)
			return
		}
		stationsArr = data.SearchByCoordinates(data.DB, lat, lon)
	}

	mainContentTemplate, err := template.ParseFiles("templates/radiostations.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	var radioContent []data.RadioContent
	for _, station := range stationsArr {
		contentType, err := data.GetContentType(station.URLResolved)
		if err != nil {
			fmt.Println(err)
			continue
		}
		radioContent = append(radioContent, data.RadioContent{Station: station, ContentType: contentType})
	}

	mainContent := data.MainContent{
		Radios: radioContent,
		GoogleMapsApiKey: credentials.GoogleMapsApiKey,
		Location: location,
	}

	err = mainContentTemplate.Execute(w, mainContent)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return
	}
}