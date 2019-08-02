package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
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

type Query struct {
	idx     int
	request string
}

var pid int
var authorid int
var userid int
var quadrantLatSize float64 = -1
var quadrantLonSize float64 = -1
var totalInstances float64 = 0
var sideLength int

func newPool(port int) *redis.Pool {

	portString := fmt.Sprintf(":%d", port)

	return &redis.Pool{
		//Max number of idle connections in the pool
		MaxIdle: 80,
		//Max number of connections
		MaxActive: 12000,

		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", portString)
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

// returns dbport, and redis ports.
// if empty, exit
func getPorts() (int, []int) {

	// Read in db port

	dbFile, err := os.Open("../ports/db.txt")
	if err != nil {
		fmt.Println(err)
		fmt.Println("ERROR: no db port")
		os.Exit(1)
	}
	scanner := bufio.NewScanner(dbFile)

	dbPort := 0

	for scanner.Scan() {
		if dbPort != 0 {
			break
		}
		line := scanner.Text()
		dbPort, err = strconv.Atoi(line)

		if err != nil {
			fmt.Println("ERROR: could not parse string in file")
			os.Exit(1)
		}
	}

	if dbPort == 0 {
		fmt.Println("ERROR: No port listening for db")
		os.Exit(1)
	}

	// Read in redis ports

	redisFile, err := os.Open("../ports/redis.txt")

	if err != nil {
		fmt.Println("ERROR: no redis port")
		os.Exit(1)
	}

	scanner = bufio.NewScanner(redisFile)

	redisPorts := make([]int, 0)

	for scanner.Scan() {
		line := scanner.Text()
		port, err := strconv.Atoi(line)

		if err != nil {
			fmt.Println("ERROR: could not parse string in file")
			os.Exit(1)
		}

		redisPorts = append(redisPorts, port)
	}

	if len(redisPorts) < 1 {
		fmt.Println("ERROR: no redis ports")
		os.Exit(1)
	}

	return dbPort, redisPorts

}

func createPools(ports []int) []*redis.Pool {
	redisInstances := make([]*redis.Pool, len(ports))

	for i := 0; i < len(ports); i++ {
		redisInstances[i] = newPool(ports[i])
	}

	// just checking if connected to all instances
	for i := 0; i < len(redisInstances); i++ {
		conn := redisInstances[i].Get()

		ping(conn)

		conn.Close()
	}

	return redisInstances
}

func powerOf4(instances int) bool {
	if instances == 0 {
		return false
	}
	for instances != 1 {
		if instances%4 != 0 {
			return false
		}
		instances = instances / 4
	}
	return true
}

// returns array index of redis index in user's region
// returns -1 if failure
func findInstanceInRadius(lat float64, lon float64) (int, int) {
	var xIndex = -1
	var yIndex = -1
	for i := 0; i < sideLength; i++ {
		if lat >= (-85 + (quadrantLatSize * float64(i))) {
			yIndex = i
		}
	}

	for i := 0; i < sideLength; i++ {
		if lon >= (-180 + (quadrantLonSize * float64(i))) {
			xIndex = i
		}
	}

	return yIndex, xIndex
}

/*func readRequest(lon float64, lat float64, redisInstances []*redis.Pool) []Note {

	yIndex, xIndex := findInstancesInRadius(lat, lon)

	reply := GEORADIUS(conn, "mapNotes", lon, lat, "4", "km")

}*/

//returns Query with {-1,""} if empty
func dequeue(queue []Query) ([]Query, Query) {
	if len(queue) <= 0 {
		return queue, Query{-1, ""}
	}
	//peek top
	x := queue[0]
	//delete top
	queue = queue[1:]

	return queue, x
}

func makeQueues() [][][]Query {
	redisQueues := make([][][]Query, sideLength)
	for i := range redisQueues {
		redisQueues[i] = make([][]Query, sideLength)
		for j := range redisQueues[i] {
			redisQueues[i][j] = make([]Query, 0)
		}
	}
	return redisQueues
}

//initialises redis instances and sets global variables
func initPools() ([]*redis.Pool, error) {

	dbPort, redisPorts := getPorts()

	fmt.Printf("dbPort: %d\n", dbPort)
	fmt.Printf("redisPorts: %v\n", redisPorts)

	//check if power of 4
	if !powerOf4(len(redisPorts)) {
		errorString := fmt.Sprintf("ERROR: Can't split %d instances into even quadrants. Need 4^n instances\n", len(redisPorts))
		return nil, errors.New(errorString)
	}

	redisInstances := createPools(redisPorts)

	//creating quadrants based on number of redis instances
	totalLat := float64(85 + 85)
	totalLon := float64(180 + 180)

	sideLength = int(math.Sqrt(float64(len(redisInstances))))
	totalInstances = float64(len(redisInstances))
	quadrantLatSize = totalLat / float64(sideLength)
	quadrantLonSize = totalLon / float64(sideLength)

	fmt.Printf("lat: %f, lon: %f per quadrant\n", quadrantLatSize, quadrantLonSize)

	return redisInstances, nil

}

func main() {

	redisInstances, err := initPools()
	if err != nil {
		fmt.Println(err)
		return
	}

	redisQueues := makeQueues()

	fmt.Println(redisInstances)
	fmt.Println(redisQueues)

	// Loop infinitely while waiting for incoming connections.
	// Decides based on coordinates which cache should be used.

	/*
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

		//delete everything on the instance
		conn.Do("FLUSHALL")*/

}
