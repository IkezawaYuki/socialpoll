package main

import (
	"flag"
	"fmt"
	"github.com/nsqio/go-nsq"
	"gopkg.in/mgo.v2"
	"log"
	"os"
	"sync"
)

var fatalErr error
func fatal(e error){
	fmt.Println(e)
	flag.PrintDefaults()
	fatalErr = e
}

func main(){
	var countsLock sync.Mutex
	var counts map[string]int

	defer func() {
		if fatalErr != nil{
			os.Exit(1)
		}
	}()
	log.Println("データベースに接続します...")
	db, err := mgo.Dial("localhost")
	if err != nil{
		fatal(err)
		return
	}
	defer func() {
		log.Println("データベース接続を閉じます...")
		db.Close()
	}()
	pollData := db.DB("ballots").C("polls")

	log.Println("NSQに接続します...")
	q, err := nsq.NewConsumer("votes", "counter", nsq.NewConfig())

	q.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		countsLock.Lock()
		defer countsLock.Unlock()
		if counts == nil{
			counts = make(map[string]int)
		}
		vote := string(m.Body)
		counts[vote]++
		return nil
	}))
	if err := q.ConnectToNSQLookupd("localhost:4161"); err != nil{
		fatal(err)
		return
	}
}