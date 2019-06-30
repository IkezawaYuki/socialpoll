package main

import (
	"github.com/bitly/go-nsq"
	"gopkg.in/mgo.v2"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var db *mgo.Session
func dialdb()error{
	var err error
	log.Println("MongoDBにダイヤル中：localhost")
	db, err = mgo.Dial("localhost")
	return err
}
func closedb(){
	db.Close()
	log.Println("データベース接続が閉じられました")
}

type poll struct{
	Options []string
}

func loadOptions()([]string, error){
	var options []string
	iter := db.DB("ballots").C("polls").Find(nil).Iter()
	var p poll
	for iter.Next(&p){
		options = append(options, p.Options...)
	}
	iter.Close()
	return options, iter.Err()
}

func publishVotes(votes <-chan string) <-chan struct{}{
	stopchan := make(chan struct{}, 1)
	pub, _ := nsq.NewProducer("localhost:4150", nsq.NewConfig())
	go func() {
		for vote := range votes{
			pub.Publish("votes", []byte(vote))
		}
		log.Println("Publisher: 停止中です。")
		pub.Stop()
		log.Println("Publisher: 停止しました。")
		stopchan <- struct{}{}
	}()
	return stopchan
}

func main() {
	if err := dialdb(); err != nil {
		log.Fatalln("MongoDBへのダイヤルに失敗しました：", err)
	}
	defer closedb()

	var stoplock sync.Mutex
	stop := false
	stopChan := make(chan struct{}, 1)
	signalChan := make(chan os.Signal, 1)
	go func() {
		<- signalChan
		stoplock.Lock()
		stop = true
		stoplock.Unlock()
		log.Println("停止します...")
		stopChan <- struct{}{}
		closeConn()
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	votes := make(chan string)
	publisherStoppedChan := publishVotes(votes)
	twitterStoppedChan := startTweeterStream(stopChan, votes)
	go func() {
		for{
			time.Sleep(1 * time.Minute)
			closeConn()
			stoplock.Lock()
			if stop{
				stoplock.Unlock()
				break
			}
			stoplock.Unlock()
		}
	}()
	<- twitterStoppedChan
	close(votes)
	<- publisherStoppedChan
}
