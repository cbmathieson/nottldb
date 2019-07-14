package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gomodule/redigo/redis"
)

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

func getRandomCoordinates() {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// get random decimal
	latDec := rand.Float64()
	longDec := rand.Float64()

	// get if negative ( 1 => -, 0 => +)

	latNeg := r.Intn(1)
	longNeg := r.Intn(1)

	// get integer

	latVal := r.Intn(84)
	longVal := r.Intn(179)

	var actualLat float64
	var actualLong float64

	if latNeg == 1 {
		actualLat = float64(latVal*(-1)) + latDec
	} else {
		actualLat = float64(latVal) + latDec
	}

	if longNeg == 1 {
		actualLong = float64(longVal*(-1)) + longDec
	} else {
		actualLong = float64(longVal) + longDec
	}

	fmt.Printf("long: %f, lat: %f", actualLong, actualLat)

}

func geoAdd(c redis.Conn) error {
	_, err := c.Do("GEOADD", "Favor")
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

func main() {

	pool := newPool()
	conn := pool.Get()
	defer conn.Close()

	err := ping(conn)
	if err != nil {
		fmt.Println(err)
	}
}
