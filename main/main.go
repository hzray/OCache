package main

import (
	"flag"
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

func createGroup() *ocache.Group {
	return ocache.NewGroup("soures", 2<<10, 2, 30, ocache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// startCacheServer 启动缓存服务器：创建 HTTPPool, 添加节点信息，注册到o中，启动HTTP服务
func startCacheServer(addr string, addrs []string, o *ocache.Group) {
	peers := ocache.NewHTTPPool(addr)
	peers.Set(addrs...)
	o.RegisterPeers(peers)
	log.Println("ocache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

// startAPIServer 用来启动一个API服务，与用户进行交互，用户感知
func startAPIServer(apiAddr string, o *ocache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := o.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("fonted server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "oCache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	o := createGroup()
	if api {
		go startAPIServer(apiAddr, o)
	}
	startCacheServer(addrMap[port], []string(addrs), o)
}
