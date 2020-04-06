package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

type poll struct {
	ID      bson.ObjectId  `bson:"id" json:"id"`
	Title   string         `json:"title"`
	Options []string       `json:"options"`
	Results map[string]int `json:"results,omitempty"`
}

func handlePolls(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method)
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
	respondErr(w, r, http.StatusNotFound)
}

func handlePollsGet(w http.ResponseWriter, r *http.Request) {
	db := GetVar(r, "db").(*mgo.Database)
	c := db.C("poll")
	var q *mgo.Query
	p := NewPath(r.URL.Path)
	if p.HasID() {
		q = c.FindId(bson.ObjectIdHex(p.ID))
	} else {
		q = c.Find(nil)
	}
	var result []*poll
	if err := q.All(&result); err != nil {
		respondErr(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, &result)
}

func handlePollsPost(w http.ResponseWriter, r *http.Request) {
	db := GetVar(r, "db").(*mgo.Database)
	c := db.C("poll")
	var p poll
	if err := decodeBody(r, &p); err != nil {
		respondErr(w, r, http.StatusBadRequest, "リクエストから調査項目を読み込めません", err)
		return
	}
	p.ID = bson.NewObjectId()
	if err := c.Insert(p); err != nil {
		respondErr(w, r, http.StatusInternalServerError, "調査項目の格納に失敗しました", err)
	}
	w.Header().Set("Location", "polls/"+p.ID.Hex())
	respond(w, r, http.StatusCreated, nil)
}

func handlePollsDelete(w http.ResponseWriter, r *http.Request) {
	db := GetVar(r, "db").(*mgo.Database)
	c := db.C("poll")
	p := NewPath(r.URL.Path)
	if !p.HasID() {
		respondErr(w, r, http.StatusMethodNotAllowed, "全ての調査項目を削除することはできません")
		return
	}
	if err := c.RemoveId(bson.ObjectIdHex(p.ID)); err != nil {
		respondErr(w, r, http.StatusInternalServerError, "調査項目の格納に失敗しました")
		return
	}
	respond(w, r, http.StatusOK, nil)
}
