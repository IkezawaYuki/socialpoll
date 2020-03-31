package main

import (
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/garyburd/go-oauth/oauth"
	"github.com/joho/godotenv"
)

var conn net.Conn
func dial(netw, addr string) (net.Conn, error){
	if conn != nil{
		conn.Close()
		conn = nil
	}
	netc, err := net.DialTimeout(netw, addr, 5*time.Second)
	if err != nil{
		return nil, err
	}
	conn = netc
	return netc, nil
}

var reader io.ReadCloser
func closeConn(){
	if conn != nil{
		conn.Close()
	}
	if reader != nil{
		reader.Close()
	}
}

var (
	authClient *oauth.Client
	creds *oauth.Credentials
)

func setupTwitterAuth(){
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	err := godotenv.Load()
	if err != nil{
		log.Fatalf("error godotenv :%v", err)
	}
	creds = &oauth.Credentials{
		Token:  os.Getenv("ACCESS_TOKEN"),
		Secret: os.Getenv("ACCESS_TOKEN_SECRET"),
	}
	authClient = &oauth.Client{
		Credentials:  oauth.Credentials{
			Token:  os.Getenv("API_KEY"),
			Secret: os.Getenv("API_SECRET_KEY"),
		},
	}
}