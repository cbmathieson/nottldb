/*

This guy is going to accept connections and toss them on backgroud
threads until completion

*/

package main

import (
	"log"
	"time"

	"github.com/mediocregopher/mediocre-go-lib/mrand"
	"github.com/mediocregopher/radix"
)

/*func uploadNote(id string) {

	var returnVal string
	p := radix.Pipeline(
		radix.FlatCmd(nil, "SET", id, 1),
		radix.Cmd(&returnVal, "GET", id),
	)

	if err := client.Do(p); err != nil {
		fmt.Println("failed to upload note")
	} else {
		fmt.Println(returnVal)
	}

	if err = client.Close(); err != nil {
		fmt.Println("failed to close client connection")
	}

}*/

const parallel = 500
const randWait = 10 // ms
func main() {
	redisPool, err := radix.NewPool("tcp", "localhost:6379", 10000)
	if err != nil {
		log.Fatalf("Redis connection error: %v\n", err)
	}
	log.Printf("Redis connection redisPool created successfully.\n")
	go func() {
		for range time.Tick(1 * time.Second) {
			log.Printf("avail conns: %d", redisPool.NumAvailConns())
		}
	}()
	log.Printf("clearing db")
	if err := redisPool.Do(radix.Cmd(nil, "FLUSHDB")); err != nil {
		log.Fatal(err)
	}
	log.Printf("populating keys")
	for i := 0; i < 100; i++ {
		err := redisPool.Do(radix.Cmd(nil, "SET", mrand.Hex(2), mrand.Hex(64)))
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("load testing")
	for i := 0; i < 10000; i++ {
		go func() {
			for {
				key := mrand.Hex(2)
				var rcv string
				err := redisPool.Do(radix.Cmd(&rcv, "GET", key))
				if err != nil {
					log.Fatalf("error getting key %q: %s", key, err)
				}
				if randWait > 0 {
					time.Sleep(time.Duration(mrand.Intn(randWait)) * time.Millisecond)
				}
			}
		}()
	}
	select {}
}
