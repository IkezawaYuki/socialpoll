package main

import (
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

type poll struct{
	ID bson.ObjectId `bson:"_id" json:"id"`
	Title string `json:"title"`
	Options []string `json:"options"`
	Results map[string]int `json:"results, omitempty"`
}


func handlePolls(w http.ResponseWriter, r *http.Request){
	switch r.Method {
	case "GET":
		handlePollsGet(w, r)
		return
	case "POST":
		handlePollsPost(w, r)
		return
	case "DELETE":
		handlePollsDelete(w, r)
		return
	case "OPTIONS":
		w.Header().Add("Access-Control-Allow-Methods", "DELETE")
		respond(w, r, http.StatusOK, nil)
		return
	}

	respondHTTPErr(w, r, http.StatusNotFound)
}

func handlePollsDelete(w http.ResponseWriter, r *http.Request) {
	db := GetVar(r, "db").(*mgo.Database)
	c := db.C("polls")
	p := NewPath(r.URL.Path)
	if !p.HasID(){
		respondErr(w, r, http.StatusMethodNotAllowed, "すべての調査項目を削除することはできません")
		return
	}
	if err := c.RemoveId(bson.ObjectIdHex(p.ID)); err != nil{
		respondErr(w, r, http.StatusInternalServerError, "調査項目の削除に失敗しました", err)
		return
	}
	respond(w, r, http.StatusOK, nil)
}


func handlePollsGet(writer http.ResponseWriter, request *http.Request) {
	db := GetVar(r, "db").(*mgo.Database)
	c := db.C("polls")
	var q *mgo.Query

	p := NewPath(request.URL.Path)
	if p.HasID(){
		q = c.FindId(bson.ObjectIdHex(p.ID))
	}else{
		q = c.Find(nil)
	}
	var result []*poll
	if err := q.All(&result); err != nil{
		respondErr(writer, request, http.StatusInternalServerError, err)
		return
	}
	respond(writer, request, http.StatusOK, &result)
}

func handlePollsPost(w http.ResponseWriter, r *http.Request){
	db := GetVar(r, "db").(*mgo.Database)
	c := db.C("polls")
	var p poll
	if err := decodeBody(r, &p); err != nil{
		respondErr(w, r, http.StatusBadRequest, "リクエストから調査項目を読み込めません", err)
		return
	}
	p.ID = bson.NewObjectId()
	if err := c.Insert(p); err != nil{
		respondErr(w, r, http.StatusInternalServerError, "調査項目の格納に失敗しました")
		return
	}
	w.Header().Set("Location", "polls/"+p.ID.Hex())
	respond(w, r, http.StatusCreated, nil)
}
















