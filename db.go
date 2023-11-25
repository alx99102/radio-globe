package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./radio.db")
	if err != nil {
		log.Fatal(err)
	}

	// create table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Radios ( id INTEGER PRIMARY KEY, name TEXT, url TEXT, url_resolved TEXT, homepage TEXT, favicon TEXT, country TEXT, language TEXT, codec TEXT, geo_lat REAL, geo_long REAL);`)
	if err != nil {
		log.Fatal(err)
	}

	// create an R-Tree index for the geo_lat and geo_long columns
	_, err = db.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS geo_index USING rtree(id, minLat, maxLat, minLon, maxLon);")
	if err != nil {
    	log.Fatal(err)
	}

	return db
}

type Station struct {
	Name string `json:"name"`
	URL string `json:"url"`
	URLResolved string `json:"url_resolved"`
	Homepage string `json:"homepage"`
	Favicon string `json:"favicon"`
	Country string `json:"country"`
	Language string `json:"language"`
	Codec string `json:"codec"`
	GeoLat float64 `json:"geo_lat"`
	GeoLong float64 `json:"geo_long"`
}

func searchByCoordinates(db *sql.DB, lat, long float64) []Station {
	rows, err := db.Query("SELECT * FROM Radios WHERE geo_lat BETWEEN ? AND ? AND geo_long BETWEEN ? AND ?", lat - 0.5, lat + 0.5, long - 0.5, long + 0.5)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var stations []Station
	var id int
	for rows.Next() {
		var station Station
		if err := rows.Scan(&id, &station.Name, &station.URL, &station.URLResolved, &station.Homepage, &station.Favicon, &station.Country, &station.Language, &station.Codec, &station.GeoLat, &station.GeoLong); err != nil {
			log.Fatal(err)
		}
		stations = append(stations, station)
	}

	return stations
}

// functions below are for ingesting the database and debugging

// func ingestDB(db *sql.DB, stations []Station) {
//     stmt, err := db.Prepare(`INSERT INTO Radios (name, url, url_resolved, homepage, favicon, country, language, codec, geo_lat, geo_long) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
//     if err != nil {
//         log.Fatal(err)
//     }
//     defer stmt.Close()

//     for _, station := range stations {
//         _, err := stmt.Exec(station.Name, station.URL, station.URLResolved, station.Homepage, station.Favicon, station.Country, station.Language, station.Codec, station.GeoLat, station.GeoLong)
//         if err != nil {
//             log.Println("Failed to insert station:", station.Name, "Error:", err)
//         }
//     }
// }


// type result struct {
	//     name       string
	// 	fileType   string
	//     err        error
	// }
	
var client = &http.Client{Timeout: 20 * time.Second}

func getContentType(URL string) (string, error) {
    // Make a request to the URL
    resp, err := client.Get(URL)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

	// Check the content-type header
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		return "", fmt.Errorf("can't determine file type")
	}

	return contentType, nil 
}

// func listRadios(db *sql.DB) string {
//     rows, err := db.Query("SELECT name, url FROM Radios")
//     if err != nil {
//         log.Fatal(err)
//     }
//     defer rows.Close()

//     var wg sync.WaitGroup
//     statusCodeChan := make(chan result)
//     errorChan := make(chan result)
//     fileTypeChan := make(chan result)

//     for rows.Next() {
//         var name, url string
//         if err := rows.Scan(&name, &url); err != nil {
//             log.Fatal(err)
//         }

//         wg.Add(1)
//         go func(name, url string) {
//             defer wg.Done()
//             var res result
//             res.name = name

//             fileType, err := getFileExtension(url)
//             if err != nil {
//                 res.err = err
//                 errorChan <- res
//                 return
//             }
//             res.fileType = fileType
//             fileTypeChan <- res
//         }(name, url)
//     }

//     go func() {
//         wg.Wait()
//         close(statusCodeChan)
//         close(errorChan)
//         close(fileTypeChan)
//     }()

//     fileTypes := make(map[string]int)
//     errorCount := 0

//     for {
//         select {
//         case res, ok := <-errorChan:
//             if !ok {
//                 err = rows.Err()
//                 if err != nil {
//                     log.Fatal(err)
//                 }
//                 // Output the file type counts and error count
//                 for fileType, count := range fileTypes {
//                     log.Println(fileType+":", count)
//                 }
//                 log.Println("Errors:", errorCount)
//                 return "File Types: " + fmt.Sprintf("%v", fileTypes) + " Errors:" + fmt.Sprint(errorCount)
//             }
//             if res.err != nil {
//                 errorCount++
//                 log.Println(res.name, "Error: ", res.err)
//             }

//         case res := <-fileTypeChan:
//             fileTypes[res.fileType]++
//         }
//     }
// }