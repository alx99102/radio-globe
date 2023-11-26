package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"radio-globe/credentials"
)

func Geocode(address string) (float64, float64, error) {
	// get API key
	creds, err := credentials.GetCredentials()
	if err != nil {
		fmt.Println("Error getting credentials:", err)
		return 0, 0, err
	}
	apiKey := creds.HereApiKey
	// build URL

	// URL encode address
	address = url.QueryEscape(address)
	
	url := "https://geocode.search.hereapi.com/v1/geocode?q=" + address + "&apiKey=" + apiKey
	
	// GET request
	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return 0, 0, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		fmt.Printf("API request failed with status code: %d\n", res.StatusCode)
		return 0, 0, fmt.Errorf("API request failed with status code: %d", res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return 0, 0, err
	}

	// parse JSON
	var hereResponse HereResponse
	err = json.Unmarshal(body, &hereResponse)
	if err != nil {
		fmt.Println(err)
		return 0, 0, err
	}
	// return lat, lng
	if len(hereResponse.Items) > 0 {
		return hereResponse.Items[0].Position.Lat, hereResponse.Items[0].Position.Lng, nil
	}
	// no results found
	return 0, 0, fmt.Errorf("no results found")
}
