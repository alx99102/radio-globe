package main

import (
	"database/sql"
	"log"

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

// var client = &http.Client{Timeout: 20 * time.Second}

// type result struct {
//     name       string
//     statusCode int
//     err        error
// }

// func listRadios(db *sql.DB) string{
//     rows, err := db.Query("SELECT name, url FROM Radios")
//     if err != nil {
//         log.Fatal(err)
//     }
//     defer rows.Close()

//     var wg sync.WaitGroup
//     statusCodeChan := make(chan result)
//     errorChan := make(chan result)

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

//             req, err := http.NewRequest("GET", url, nil)
//             if err != nil {
//                 res.err = err
//                 errorChan <- res
//                 return
//             }

//             response, err := client.Do(req)
//             if err != nil {
//                 res.err = err
//                 errorChan <- res
//                 return
//             }
//             res.statusCode = response.StatusCode
//             statusCodeChan <- res
//         }(name, url)
//     }

//     go func() {
//         wg.Wait()
//         close(statusCodeChan)
//         close(errorChan)
//     }()

//     code200s := 0
//     code400s := 0
//     code500s := 0
// 	errorCount := 0

//     for {
//         select {
//         case res, ok := <-errorChan:
//             if !ok {
//                 err = rows.Err()
//                 if err != nil {
//                     log.Fatal(err)
//                 }
//                 log.Println("200s:", code200s, " 400s:", code400s, " 500s:", code500s, " Errors:", errorCount)
// 				return "Concurrent: 200s:" + string(code200s) + " 400s:" + string(code400s) + " 500s:" + string(code500s) + " Errors:" + string(errorCount)
//             }
//             if res.err != nil {
// 				errorCount++
//                 log.Println(res.name, "Error: ", res.err)
//             }

//         case res, ok := <-statusCodeChan:
//             if !ok {
//                 continue
//             }
//             log.Println(res.name, res.statusCode)
//             switch res.statusCode {
//             case 200:
//                 code200s++
//             case 400:
//                 code400s++
//             case 500:
//                 code500s++
//             }
//         }
//     }
// }