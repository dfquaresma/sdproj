package main

import (
	"fmt"
	"github.com/gci-proxy-resolver/function"
	"github.com/gci-proxy-resolver/model"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	// TODO: We can do by implementing a network BROADCAST to know the cluster managers addresses.
	managers := make([]string, 0, 10)
	for _, managerAddress := range strings.Split(os.Getenv("MANAGER_ADDRESSES"), ",") {
		managers = append(managers, strings.TrimSpace(managerAddress))
	}

	router := mux.NewRouter()
	router.HandleFunc("/", handle()).Methods(http.MethodGet, http.MethodPost)

	s := &http.Server{
		Addr:           "127.0.0.1:8082",
		Handler:        router,
	}

	go getNodesList()
	go getPublishedPorts()

	log.Fatal(s.ListenAndServe())
}

func getNodesList() {
	for {

	}
}

func getPublishedPorts() {

}

func handle() http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Cannot read request body %+v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Cannot read request body %+v\n", err)
			return
		}
		headers := http.Header{}
		for key, value := range r.Header {
			if !validHeader(key) {
				for _, v := range value {
					headers.Add(key, v)
				}
			}
		}
		resp, err := function.Handle(model.Request{Body: body, Headers: headers})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error on function handle call %+v\n", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(resp.Body)
	}
}

func validHeader(key string) bool {
	return false
}