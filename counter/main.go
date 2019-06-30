package main

import (
	"flag"
	"fmt"
	"github.com/bitly/go-nsq"
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
	defer func(){
		if fatalErr != nil{
			os.Exit(1)
		}
	}()
	log.Println("データベースに接続します...")
	db, err := mgo.Dial("localhost")
	if err != nil {
		fatal(err)
	}
	defer func() {
		log.Println("データベースの接続を閉じます...")
		db.Close()
	}()
	pollData := db.DB("ballots").C("polls")

	var countsLock sync.Mutex
	var counts map[string]int

	log.Println("NSQに接続します...")
	q, err := nsq.NewConsumer("votes", "counter", nsq.NewConfig())
	if err != nil {
		fatal(err)
		return
	}

	q.AddHandler(nsq.HandlerFunc(func(m *nsq.Message)error {
		//todo

		return nil
	}))
}