package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	redis "gopkg.in/redis.v4"
)

var configurationTable = "configuration"

type configuration struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type dbinterface struct {
	client *redis.Client
}

func main() {
	db := &dbinterface{client: redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})}

	_, err := db.client.Ping().Result()
	if err != nil {
		fmt.Println("Db connection KO", os.Getenv("REDIS_URL"))
		return
	}
	fmt.Println("Db connection OK", os.Getenv("REDIS_URL"))

	router := mux.NewRouter()
	router.HandleFunc("/{id}", db.getHandler).Methods("GET")
	router.HandleFunc("/{id}", db.postHandler).Methods("POST")
	router.HandleFunc("/{id}", db.putHandler).Methods("PUT")
	router.HandleFunc("/{id}", db.deleteHandler).Methods("DELETE")
	router.HandleFunc("/", db.getAllHandler).Methods("GET")
	router.Use(mdw)

	port := os.Getenv("APP_PORT")
	fmt.Println("Listening on port", port)

	http.ListenAndServe(":"+port, router)
}

func mdw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json; charset=UTF-8")
		next.ServeHTTP(w, r)
	})
}

func (db *dbinterface) getHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	// Read config from db
	val, err := db.client.HGet(configurationTable, id).Result()
	if err == redis.Nil {
		log.Println("Error ", err)
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		log.Print("Error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Write config to response
	_, err = w.Write([]byte(val))
	if err != nil {
		log.Print("Error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (db *dbinterface) postHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	// Decode config data
	config := configuration{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		log.Println("Error parsing JSON", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate config
	if !isValid(id, config) {
		log.Print("Configuration is not valid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Verify that config id does not exist yet
	_, err := db.client.HGet(configurationTable, id).Result()
	if err != redis.Nil {
		log.Println("Error id "+id+" already exists", err)
		w.WriteHeader(http.StatusConflict)
		return
	}

	// Serialize to JSON
	data, err := json.Marshal(config)
	if err != nil {
		log.Print("Error marshalling JSON", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Store in db
	err = db.client.HSet(configurationTable, id, string(data)).Err()
	if err != nil {
		log.Print("Error writing data", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (db *dbinterface) putHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	// Parse from JSON
	config := configuration{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		log.Println("Error parsing JSON", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate the config data
	if !isValid(id, config) {
		log.Print("Configuration is not valid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Verify that the config id exists
	_, err := db.client.HGet(configurationTable, id).Result()
	if err != nil {
		log.Println("Error retrieving config with id "+id, err)
		w.WriteHeader(http.StatusConflict)
		return
	}

	// Serialize config data to JSON
	data, err := json.Marshal(config)
	if err != nil {
		log.Print("Error marshalling JSON", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Store in db
	err = db.client.HSet(configurationTable, id, string(data)).Err()
	if err != nil {
		log.Print("Error writing data", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (db *dbinterface) deleteHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	db.client.HDel(configurationTable, id)
}

func (db *dbinterface) getAllHandler(w http.ResponseWriter, r *http.Request) {

	// Read all key, value pairs in configuration table
	configs, err := db.client.HGetAll(configurationTable).Result()
	if err != nil {
		log.Print("Error retrieving configs", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Serialize to JSON
	data, err := json.Marshal(configs)
	if err != nil {
		log.Print("Error marshalling JSON", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Write to response
	_, err = w.Write([]byte(data))
	if err != nil {
		log.Print("Error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func isValid(id string, config configuration) bool {
	if config.ID == id {
		return true
	}

	return false
}
