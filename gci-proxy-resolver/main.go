package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dfquaresma/sdproj/gci-proxy-resolver/model"
	"github.com/docker/docker/api/types/swarm"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	// TODO: We can do it by implementing a network BROADCAST to know the cluster managers addresses.
	managers := make([]string, 0, 10)
	for _, managerAddress := range strings.Split(os.Getenv("MANAGER_ADDRESSES"), ",") {
		managers = append(managers, strings.TrimSpace(managerAddress))
	}

	clusterInfo := model.ClusterInfo{ManagerAddresses:managers}

	router := mux.NewRouter()
	router.HandleFunc("/", handle(&clusterInfo)).Methods(http.MethodGet)

	s := &http.Server{
		Addr:           "127.0.0.1:8082",
		Handler:        router,
	}

	conciliationTime := 5 * time.Second
	apiVersion := "v1.40"
	go updateInfo(&clusterInfo, conciliationTime, apiVersion)

	log.Fatal(s.ListenAndServe())
}

func updateInfo(clusterInfo *model.ClusterInfo, conciliationTime time.Duration, apiVersion string) {
	n := len(clusterInfo.ManagerAddresses)
	for i := 0;; i = (i + 1) % n {
		log.Printf("Using manager %s", clusterInfo.ManagerAddresses[i])

		nodesUrl := fmt.Sprintf("http://%s/%s/nodes", clusterInfo.ManagerAddresses[i], apiVersion)
		nodeList, err := getNodesList(nodesUrl, strings.Split(clusterInfo.ManagerAddresses[i], ":")[0])
		if err == nil {
			log.Printf("Updating nodes list to: %+q", nodeList)
			clusterInfo.NodeIPs = nodeList
		} else {
			log.Printf("Cannot update nodes list due to: %s", err.Error())
		}
		time.Sleep(conciliationTime)
	}
}

func getNodesList(url string, managerIP string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Received http status code %d and error while trying to read body: %v", resp.StatusCode, err))
	}

	var nodeIPs []string
	if resp.StatusCode < 300 {
		var data []swarm.Node

		err = json.Unmarshal(body, &data)
		if err == nil {
			nodeIPs = make([]string, 0, 10)
			for _, node := range data {
				if node.Status.State == swarm.NodeStateReady {
					if node.Status.Addr == "0.0.0.0" {
						nodeIPs = append(nodeIPs, managerIP)
					} else {
						nodeIPs = append(nodeIPs, node.Status.Addr)
					}
				}
			}
		}
	} else {
		err = errors.New(fmt.Sprintf("Received http status code %d and response body: %s", resp.StatusCode, body))
	}
	return nodeIPs, err
}

func handle(info *model.ClusterInfo) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ioutil.ReadAll(r.Body)
		resp, err := json.Marshal(info)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Could not serialize json due to: %s\n", err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}
