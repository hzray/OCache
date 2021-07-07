package main

import (
	"fmt"
	"log"
	"net/http"
	"ocache"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	ocache.NewGroup("scores", 2<<10, ocache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := ocache.NewHTTPPool(addr)
	log.Println("ocache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
