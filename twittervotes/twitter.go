package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
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


var (
	authSetupOnce sync.Once
	httpClient *http.Client
)

func makeRequest(req *http.Request, params url.Values) (*http.Response, error){
	authSetupOnce.Do(func(){
		setupTwitterAuth()
		httpClient = &http.Client{
			Transport:     &http.Transport{
				Dial: dial,
			},
		}
	})
	formEnc := params.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(formEnc)))
	req.Header.Set("Authorization", authClient.AuthorizationHeader(creds, "POST", req.URL, params))
	return httpClient.Do(req)
}

type tweet struct {
	Text string
}

func readFromTwitter(votes chan<- string){
	options, err := loadOptions()
	if err != nil{
		log.Println("選択肢の読み込みに失敗しました：", err)
		return
	}
	u, err := url.Parse("https://stream.twitter.com/1.1/statuses/filter.json")
	if err != nil {
		log.Println("検索のリクエストに失敗しました：", err)
		return
	}
	query := make(url.Values)
	query.Set("track", strings.Join(options, ","))
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(query.Encode()))
	resp, err := makeRequest(req, query)
	if err != nil{
		log.Println("検索のリクエストに失敗しました：", err)
		return
	}
	reader = resp.Body
	decoder := json.NewDecoder(reader)
	for{
		var tweet tweet
		if err := decoder.Decode(&tweet); err != nil{
			break
		}
		for _, option := range options{
			if strings.Contains(strings.ToLower(tweet.Text), strings.ToLower(option)){
				log.Println("投票：", option)
				votes <- option
			}
		}
	}
}

func startTwitterStream(stopchan <-chan struct{}, votes chan<- string)<-chan struct{}{
	stoppedchan := make(chan struct{}, 1)
	go func(){
		defer func() {
			stoppedchan <- struct{}{}
		}()
		for {
			select{
			case <- stopchan:
				log.Println("Twitterへの問合せを終了します...")
				return
			default:
				log.Println("Twitterに問合せます...")
				readFromTwitter(votes)
				log.Println("（待機中）")
				time.Sleep(10 * time.Second)
			}
		}
	}()
	return stopchan
}