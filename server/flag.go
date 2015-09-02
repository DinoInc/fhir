package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/search"
	"gopkg.in/mgo.v2/bson"
)

func FlagIndexHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	var result []models.Flag
	c := Database.C("flags")

	r.ParseForm()
	if len(r.Form) == 0 {
		iter := c.Find(nil).Limit(100).Iter()
		err := iter.All(&result)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	} else {
		searcher := search.NewMongoSearcher(Database)
		query := search.Query{Resource: "Flag", Query: r.URL.RawQuery}
		err := searcher.CreateQuery(query).All(&result)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	}

	var flagEntryList []models.BundleEntryComponent
	for i := range result {
		var entry models.BundleEntryComponent
		entry.Resource = &result[i]
		flagEntryList = append(flagEntryList, entry)
	}

	var bundle models.Bundle
	bundle.Id = bson.NewObjectId().Hex()
	bundle.Type = "searchset"
	var total = uint32(len(result))
	bundle.Total = &total
	bundle.Entry = flagEntryList

	log.Println("Setting flag search context")
	context.Set(r, "Flag", result)
	context.Set(r, "Resource", "Flag")
	context.Set(r, "Action", "search")

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(rw).Encode(&bundle)
}

func LoadFlag(r *http.Request) (*models.Flag, error) {
	var id bson.ObjectId

	idString := mux.Vars(r)["id"]
	if bson.IsObjectIdHex(idString) {
		id = bson.ObjectIdHex(idString)
	} else {
		return nil, errors.New("Invalid id")
	}

	c := Database.C("flags")
	result := models.Flag{}
	err := c.Find(bson.M{"_id": id.Hex()}).One(&result)
	if err != nil {
		return nil, err
	}

	log.Println("Setting flag read context")
	context.Set(r, "Flag", result)
	context.Set(r, "Resource", "Flag")
	return &result, nil
}

func FlagShowHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	context.Set(r, "Action", "read")
	_, err := LoadFlag(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(rw).Encode(context.Get(r, "Flag"))
}

func FlagCreateHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	decoder := json.NewDecoder(r.Body)
	flag := &models.Flag{}
	err := decoder.Decode(flag)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	c := Database.C("flags")
	i := bson.NewObjectId()
	flag.Id = i.Hex()
	err = c.Insert(flag)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	log.Println("Setting flag create context")
	context.Set(r, "Flag", flag)
	context.Set(r, "Resource", "Flag")
	context.Set(r, "Action", "create")

	host, err := os.Hostname()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	rw.Header().Add("Location", "http://"+host+":3001/Flag/"+i.Hex())
}

func FlagUpdateHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	var id bson.ObjectId

	idString := mux.Vars(r)["id"]
	if bson.IsObjectIdHex(idString) {
		id = bson.ObjectIdHex(idString)
	} else {
		http.Error(rw, "Invalid id", http.StatusBadRequest)
	}

	decoder := json.NewDecoder(r.Body)
	flag := &models.Flag{}
	err := decoder.Decode(flag)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	c := Database.C("flags")
	flag.Id = id.Hex()
	err = c.Update(bson.M{"_id": id.Hex()}, flag)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	log.Println("Setting flag update context")
	context.Set(r, "Flag", flag)
	context.Set(r, "Resource", "Flag")
	context.Set(r, "Action", "update")
}

func FlagDeleteHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	var id bson.ObjectId

	idString := mux.Vars(r)["id"]
	if bson.IsObjectIdHex(idString) {
		id = bson.ObjectIdHex(idString)
	} else {
		http.Error(rw, "Invalid id", http.StatusBadRequest)
	}

	c := Database.C("flags")

	err := c.Remove(bson.M{"_id": id.Hex()})
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Setting flag delete context")
	context.Set(r, "Flag", id.Hex())
	context.Set(r, "Resource", "Flag")
	context.Set(r, "Action", "delete")
}