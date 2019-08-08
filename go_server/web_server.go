package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"
)

type userData struct {
	pid        int
	email      string
	profileImg string
	username   string
}

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

type Redis struct {
	Pool  *redis.Pool
	Queue []string
}

type Get struct {
	Lat float64
	Lon float64
	Id  int
}

type Post struct {
	Lat  float64
	Lon  float64
	Data Note
}

type Response struct {
	Ok bool
}

type Envelope struct {
	Type string
	Msg  interface{}
}

var pid int
var authorid int
var userid int
var quadrantLatSize float64 = -1
var quadrantLonSize float64 = -1
var sideLength int

func main() {

	// create grid of Redis objects
	redisServers, err := makeRedisServers()
	if err != nil {
		fmt.Println("ERROR: could not initialize redis servers")
		return
	}

	// connect to user database
	db, err := connectToDB()
	defer db.Close()

	// Loop infinitely while waiting for incoming connections.
	// Decides based on coordinates which cache should be used.
	for x := range redisServers {
		for y := range redisServers[x] {
			go func(server *Redis) {
				for {
					// wait for the queue to have content
					if len(server.Queue) > 0 {
						err := handleRequest(server)
						if err != nil {
							fmt.Println(err)
						}
					} else {
						time.Sleep(50 * time.Millisecond)
					}
				}
			}(&redisServers[x][y])
		}
	}

	// Randomly adds get requests to queue
	go createGetRequests(redisServers)

	// generates notes, adds post requests to queue
	go createPostRequests(redisServers)

	// FUTURE:
	// generate database write requests
	// generate database read requests

	time.Sleep(10000 * time.Second)

}

// ------------------ Request Makers/Handlers ------------------

func createPostRequests(redisServers [][]Redis) {
	for {
		time.Sleep(500 * time.Millisecond)

		// creates artificial note
		note := generateNote()

		// uses location passed by note to decide what redis instance to pass to
		x, y := findInstanceInRadius(note.Longitude, note.Latitude)

		request := Envelope{"post", Post{note.Latitude, note.Longitude, note}}
		json_request, err := json.Marshal(request)
		if err != nil {
			fmt.Println(err)
			return
		}
		// adds note to queue
		redisServers[x][y].Queue = enqueue(redisServers[x][y].Queue, string(json_request))
	}
}

func createGetRequests(redisServers [][]Redis) {
	for {

		userID := 0

		// block until recieve request
		time.Sleep(500 * time.Millisecond)

		// get artificial coordinates for a user
		userLon, userLat := getRandomCoordinates()
		x, y := findInstanceInRadius(userLon, userLat)

		request := Envelope{"get", Get{userLat, userLon, userID}}
		json_request, err := json.Marshal(request)
		if err != nil {
			fmt.Println(err)
			return
		}
		redisServers[x][y].Queue = enqueue(redisServers[x][y].Queue, string(json_request))
		userID++
	}
}

func handleRequest(server *Redis) error {
	var request string
	server.Queue, request = dequeue(server.Queue)

	//Decide if Get or Post
	var msg json.RawMessage
	env := Envelope{
		Msg: &msg,
	}

	if err := json.Unmarshal([]byte(request), &env); err != nil {
		return err
	}

	switch env.Type {
	case "get":
		fmt.Println("get request!")

		var get Get
		err := json.Unmarshal(msg, &get)
		if err != nil {
			return err
		}

		notes := readRequest(get.Lon, get.Lat, server.Pool)
		println("notes in instance: ", len(notes))
	case "post":
		fmt.Println("post request!")
		var post Post
		err := json.Unmarshal(msg, &post)
		if err != nil {
			return err
		}
		err = addNote(post.Data, server.Pool)
		if err != nil {
			return err
		}
		fmt.Println("Successfully added note!")
	default:
		fmt.Printf("unknown message type: %s\n", env.Type)
		return errors.New("ERROR: Unknown message type")
	}
	return nil
}

// --------------------------------------------------------

// ------------------ Postgres interactions ---------------

func connectToDB() (*sql.DB, error) {
	dbPort, _ := getPorts()
	psqlCreds := fmt.Sprintf("host=localhost port=%d user=postgres password=pass dbname=postgres sslmode=disable", dbPort)

	// connect to user_db
	db, err := sql.Open("postgres", psqlCreds)
	if err != nil {
		return db, err
	}

	err = db.Ping()
	if err != nil {
		return db, err
	}

	return db, err
}

func createUser(db *sql.DB, user userData) error {
	sqlStatement := `
	INSERT INTO users (pid, email, profileImg, username)
	VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(sqlStatement, strconv.Itoa(user.pid), user.email, user.profileImg, user.username)
	return err
}

func createNewNote(db *sql.DB, note Note) error {
	sqlStatement := `
	INSERT INTO notes (pid, email, profileImg, username)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := db.Exec(sqlStatement, note.Pid, note.Caption, note.DateCreated, note.DateFound, note.IsAnonymous, note.Latitude, note.Longitude, note.NoteImage, note.Author_id)
	return err
}

func addToFavorites(db *sql.DB, uid string, note Note) error {
	sqlStatement := `
	INSERT INTO users (user_id, note_id)
	VALUES ($1, $2)`
	_, err := db.Exec(sqlStatement, uid, strconv.Itoa(note.Pid))
	return err
}

func getMyNotes(db *sql.DB, uid string) ([]Note, error) {
	rows, err := db.Query("SELECT * FROM notes WHERE author_id=$1", uid)
	notes := make([]Note, 0)
	if err != nil {
		return notes, err
	}

	defer rows.Close()

	for rows.Next() {
		note := Note{}
		err = rows.Scan(&note.Pid, &note.Caption, &note.DateCreated, &note.DateFound, &note.IsAnonymous, &note.Latitude, &note.Longitude, &note.NoteImage, &note.Author_id)
		if err != nil {
			return notes, err
		}
		notes = append(notes, note)
	}

	err = rows.Err()
	return notes, err
}

func queryFavorites(db *sql.DB, uid int) ([]Note, error) {

	rows, err := db.Query("SELECT * FROM notes INNER JOIN favorites ON notes.pid=favorites.note_id AND favorites.user_id=$1", uid)
	notes := make([]Note, 0)

	if err != nil {
		return notes, err
	}

	defer rows.Close()

	for rows.Next() {
		note := Note{}
		err = rows.Scan(&note.Pid, &note.Caption, &note.DateCreated, &note.DateFound, &note.IsAnonymous, &note.Latitude, &note.Longitude, &note.NoteImage, &note.Author_id)
		if err != nil {
			return notes, err
		}
		notes = append(notes, note)
	}

	err = rows.Err()
	return notes, err
}

// --------------------------------------------------------

// ------------------ redis interactions ------------------

// adds note to map
func addNote(note Note, pool *redis.Pool) error {

	// create connection
	c := pool.Get()
	defer c.Close()

	//marshal into json
	b, err := json.Marshal(note)
	if err != nil {
		return err
	}

	//fmt.Println(string(b))
	input := string(b)

	//fmt.Printf("note lat: %f, lon: %f\n", note.Latitude, note.Longitude)

	err = GEOADD(c, note, input)

	if err != nil {
		return err
	}
	return nil
}

// generates 2D slice of redis instances based on ports given
func createPools(ports []int) [][]*redis.Pool {
	redisInstances := make([][]*redis.Pool, sideLength)

	for i := 0; i < sideLength; i++ {
		redisInstances[i] = make([]*redis.Pool, sideLength)
		for j := 0; j < sideLength; j++ {
			redisInstances[i][j] = newPool(ports[i])

			// just checking if connected to instance
			conn := redisInstances[i][j].Get()
			ping(conn)
			conn.Close()
		}
	}

	return redisInstances
}

// handles user request for notes in area
func readRequest(lon float64, lat float64, pool *redis.Pool) []Note {

	//fmt.Println(pool)

	c := pool.Get()
	defer c.Close()

	//fmt.Printf("searching 21,000km radius from lat: %f, lon: %f\n", lon, lat)

	reply := GEORADIUS(c, lon, lat, "21000", "km")

	var notes []Note

	switch t := reply.(type) {
	case []interface{}:
		returnedValues := make([]Note, len(t))
		for i, value := range t {
			if err := json.Unmarshal(value.([]byte), &returnedValues[i]); err != nil {
				panic(err)
			}
		}
		notes = returnedValues
	default:
		fmt.Println("uh oh not a data type we wanted")
	}

	return notes

}

//initialises redis instances and sets global variables
func initPools() ([][]*redis.Pool, error) {

	dbPort, redisPorts := getPorts()

	fmt.Printf("dbPort: %d\n", dbPort)
	fmt.Printf("redisPorts: %v\n", redisPorts)

	//check if power of 4
	if !powerOf4(len(redisPorts)) {
		errorString := fmt.Sprintf("ERROR: Can't split %d instances into even quadrants. Need 4^n instances\n", len(redisPorts))
		return nil, errors.New(errorString)
	}

	//creating quadrants based on number of redis instances
	totalLat := float64(85 + 85)
	totalLon := float64(180 + 180)

	sideLength = int(math.Sqrt(float64(len(redisPorts))))

	quadrantLatSize = totalLat / float64(sideLength)
	quadrantLonSize = totalLon / float64(sideLength)

	//fmt.Printf("lat: %f, lon: %f per quadrant\n", quadrantLatSize, quadrantLonSize)

	redisInstances := createPools(redisPorts)

	return redisInstances, nil

}

// ---------------------------------------------------

// ------------------ Utilities ------------------

// creates an artificial note
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
		"", //"https://firebasestorage.googleapis.com/v0/b/nottl-92731.appspot.com/o/notes%2F01DE3BD6-2807-4947-B39A-8A7F13397EE0.jpg?alt=media&token=12f02b79-894d-497e-985f-0f10063c58da",
		"", //https://firebasestorage.googleapis.com/v0/b/nottl-92731.appspot.com/o/profile_pictures%2F27171C21-C9BD-4452-AF66-DC417DD508D8.jpg?alt=media&token=4e217783-f10c-44e0-8b82-b0104d59010d",
		authorid,
	}
}

// generates random lon,lat values
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

// checks if values is a power of 4
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
// lon maps to x and lat maps to y
// returns -1 if failure
func findInstanceInRadius(lon float64, lat float64) (int, int) {
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

	return xIndex, yIndex
}

func dequeue(queue []string) ([]string, string) {
	if len(queue) == 0 {
		return queue, ""
	}
	// peek top
	x := queue[0]
	// delete top
	queue = queue[1:]

	return queue, x
}

func enqueue(queue []string, item string) []string {
	queue = append(queue, item)
	return queue
}

// creates 2D slice of queues that map to each redis instance using 'findInstanceInRadius()'
func makeRedisServers() ([][]Redis, error) {

	// generate connection pools
	redisInstances, err := initPools()
	if err != nil {
		fmt.Println(err)
		return [][]Redis{}, err
	}

	server := make([][]Redis, sideLength)

	for x := range server {
		server[x] = make([]Redis, sideLength)
		for y := 0; y < sideLength; y++ {
			server[x][y].Pool = redisInstances[x][y]
		}
	}

	return server, nil
}

// -----------------------------------------------

// ------------------ redis functions ------------------

// connects to the redis instance located at the given port
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

func GEOADD(c redis.Conn, note Note, input string) error {
	_, err := c.Do("GEOADD", "mapNotes", note.Longitude, note.Latitude, input)
	return err
}

//test connectivity
func ping(c redis.Conn) error {
	pong, err := c.Do("PING")
	if err != nil {
		return err
	}

	_, err = redis.String(pong, err)
	if err != nil {
		return err
	}

	//fmt.Printf("PING Response = %s\n", s)
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

func GEORADIUS(c redis.Conn, lon float64, lat float64, radius string, unit string) interface{} {
	resp, err := c.Do("GEORADIUS", "mapNotes", lon, lat, radius, unit)
	if err != nil {
		fmt.Println(err)
	}
	return resp
}

// -----------------------------------------------
