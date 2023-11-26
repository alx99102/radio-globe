package credentials

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Credentials struct {
	GeoapifyApiKey   string `json:"geoapify_api_key"`
	GoogleMapsApiKey string `json:"google_maps_api_key"`
	HereApiKey       string `json:"here_api_key"`
}

func GetCredentials() (Credentials, error){
	credsString, err := ioutil.ReadFile("./credentials/creds.json")
	if err != nil {
		fmt.Println(err)
		return Credentials{}, err
	}
	var credentials Credentials
	err = json.Unmarshal(credsString, &credentials)
	if err != nil {
		fmt.Println(err)
		return Credentials{}, err
	}
	return credentials, nil
}