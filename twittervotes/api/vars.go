package main

import (
	"flag"
	"github.com/stretchr/graceful"
	"gopkg.in/mgo.v2"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	varsLock sync.RWMutex
	vars     map[*http.Request]map[string]interface{}
)

func main() {
	var (
		addr  = flag.String("addr", ":8080", "エンドポイントのアドレス")
		mongo = flag.String("mongo", "localhost", "MongoDBのアドレス")
	)
	flag.Parse()
	log.Println("MongoDBに接続します", *mongo)
	db, err := mgo.Dial(*mongo)
	if err != nil {
		log.Fatalln("MongoDBへの接続に失敗しました")
	}
	defer db.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("/polls/", withCORS(withVars(withData(db, withAPIKey(handlePolls)))))
	mux.HandleFunc("/polls", withCORS(withVars(withData(db, withAPIKey(handlePolls)))))
	log.Println("Webサーバーを開始します：", *addr)
	graceful.Run(*addr, 1*time.Second, mux)
	log.Println("停止します...")
}

func withAPIKey(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isValidAPIKey(r.URL.Query().Get("key")) {
			respondErr(w, r, http.StatusUnauthorized, "不正なAPIキーです")
			return
		}
		fn(w, r)
	}
}

func withData(d *mgo.Session, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		thisDb := d.Copy()
		defer thisDb.Close()
		SetVar(r, "db", thisDb.DB("ballots"))
		f(w, r)
	}
}

func withVars(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		OpenVars(r)
		defer CloseVars(r)
		fn(w, r)
	}
}

func withCORS(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Location")
		fn(w, r)
	}
}

func isValidAPIKey(key string) bool {
	return "abc123" == key
}

func OpenVars(r *http.Request) {
	varsLock.Lock()
	if vars == nil {
		vars = map[*http.Request]map[string]interface{}{}
	}
	vars[r] = map[string]interface{}{}
	varsLock.Unlock()
}

func CloseVars(r *http.Request) {
	varsLock.Lock()
	delete(vars, r)
	varsLock.Unlock()
}

func GetVar(r *http.Request, key string) interface{} {
	varsLock.Lock()
	value := vars[r][key]
	varsLock.Unlock()
	return value
}

func SetVar(r *http.Request, key string, value interface{}) {
	varsLock.Lock()
	vars[r][key] = value
	varsLock.Unlock()
}
