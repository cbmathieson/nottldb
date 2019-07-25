package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Note struct {
	Pid          int
	Caption      string
	DateCreated  string
	DateFound    string
	IsAnonymous  bool
	Latitude     float64
	Longitude    float64
	NoteImage    string
	ProfileImage string
	Author_id    int
}

var pid int
var authorid int
var userid int

func newPool() *redis.Pool {
	return &redis.Pool{
		//Max number of idle connections in the pool
		MaxIdle: 80,
		//Max number of connections
		MaxActive: 12000,

		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

func generateNote() Note {
	lon, lat := getRandomCoordinates()
	pid++
	authorid++
	return Note{
		pid,
		"",
		"",
		"",
		false,
		lat,
		lon,
		"https://firebasestorage.googleapis.com/v0/b/nottl-92731.appspot.com/o/notes%2F01DE3BD6-2807-4947-B39A-8A7F13397EE0.jpg?alt=media&token=12f02b79-894d-497e-985f-0f10063c58da",
		"https://firebasestorage.googleapis.com/v0/b/nottl-92731.appspot.com/o/profile_pictures%2F27171C21-C9BD-4452-AF66-DC417DD508D8.jpg?alt=media&token=4e217783-f10c-44e0-8b82-b0104d59010d",
		authorid,
	}
}

func addNote(c redis.Conn) error {
	//marshal into json
	note := generateNote()
	//fmt.Println(note)
	b, err := json.Marshal(note)
	if err != nil {
		return err
	}

	//fmt.Println(string(b))
	input := string(b)

	fmt.Printf("note lat: %f, lon: %f\n", note.Latitude, note.Longitude)

	_, err = c.Do("GEOADD", "mapNotes", note.Longitude, note.Latitude, input)

	if err != nil {
		return err
	}
	return nil
}

func getRandomCoordinates() (float64, float64) {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// get random decimal
	latDec := rand.Float64()
	longDec := rand.Float64()

	// get if negative ( 2 => -, 1 => +)

	latNeg := r.Intn(2)
	longNeg := r.Intn(2)

	// get integer

	latVal := r.Intn(84)
	longVal := r.Intn(179)

	var actualLat float64
	var actualLong float64

	if latNeg == 1 {
		actualLat = float64(latVal*(-1)) - latDec
	} else {
		actualLat = float64(latVal) + latDec
	}

	if longNeg == 1 {
		actualLong = float64(longVal*(-1)) - longDec
	} else {
		actualLong = float64(longVal) + longDec
	}

	return actualLong, actualLat

}

func geoAdd(c redis.Conn) error {
	_, err := c.Do("GEOADD", "Favor")
	return err
}

//test connectivity
func ping(c redis.Conn) error {
	pong, err := c.Do("PING")
	if err != nil {
		return err
	}

	s, err := redis.String(pong, err)
	if err != nil {
		return err
	}

	fmt.Printf("PING Response = %s\n", s)
	return nil
}

func SET(c redis.Conn, key string, val interface{}) bool {
	resp, err := redis.String(c.Do("SET", key, val))
	if err != nil {
		fmt.Println(err)
	}
	return resp == "OK"
}

// Set if doesn't exist
func SETNX(c redis.Conn, key string, val interface{}) bool {
	resp, err := redis.Int(c.Do("SETNX", key, val))
	if err != nil {
		fmt.Println(err)
	}
	return resp == 1
}

func INCR(c redis.Conn, key string) interface{} {
	resp, err := c.Do("INCR", key)
	if err != nil {
		fmt.Println(err)
	}
	return resp
}

// Deletes key after time seconds
func EXPIRE(c redis.Conn, key string, time int) bool {
	resp, err := redis.Int(c.Do("EXPIRE", key, time))
	if err != nil {
		fmt.Println(err)
	}
	return resp == 1
}

// Time to live for key with expiration time
func TTL(c redis.Conn, key string) interface{} {
	resp, err := c.Do("TTL", key)
	if err != nil {
		fmt.Println(err)
	}
	return resp
}

func GET(c redis.Conn, key string) interface{} {
	resp, err := redis.String(c.Do("GET", key))
	if err != nil {
		fmt.Println(err)
	}
	return resp
}

func DEL(c redis.Conn, key string) bool {
	resp, err := redis.Int(c.Do("DEL", key))
	if err != nil {
		fmt.Println(err)
	}
	return resp == 1
}

func GEORADIUS(c redis.Conn, key string, lon float64, lat float64, radius string, unit string) interface{} {
	resp, err := c.Do("GEORADIUS", key, lon, lat, "21000", "km")
	if err != nil {
		fmt.Println(err)
	}
	return resp
}

func main() {

	pool := newPool()
	conn := pool.Get()
	defer conn.Close()

	// Adding notes to the map
	for i := 0; i < 5; i++ {

		err := addNote(conn)
		if err != nil {
			fmt.Printf("failed to add note")
			fmt.Println(err)
			break
		}

	}

	//get location to search from
	lon, lat := getRandomCoordinates()
	fmt.Printf("searching 21,000km radius from lat: %f, lon: %f\n", lon, lat)
	// search in the maximum radius on earths surface (20,905km)
	reply := GEORADIUS(conn, "mapNotes", lon, lat, "21000", "km")

	switch t := reply.(type) {
	case []interface{}:
		returnedValues := make([]Note, len(t))
		for i, value := range t {
			if err := json.Unmarshal(value.([]byte), &returnedValues[i]); err != nil {
				panic(err)
			}
			fmt.Println(returnedValues[i])
		}
	default:
		fmt.Println("uh oh not a data type we wanted\n")
	}

	conn.Do("FLUSHALL")
}
