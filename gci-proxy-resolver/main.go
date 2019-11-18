package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/swarm"
	"github.com/gorilla/mux"
	"github.com/sdproj/gci-proxy-resolver/model"
	//"github.com/docker/docker/api/types"
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
	router.HandleFunc("/", handle(&clusterInfo)).Methods(http.MethodPost)

	s := &http.Server{
		Addr:           "127.0.0.1:8082",
		Handler:        router,
	}

	conciliationTime := 10 * time.Second
	apiVersion := "v1.40"
	go updateInfo(&clusterInfo, conciliationTime, apiVersion)

	log.Fatal(s.ListenAndServe())
}

func updateInfo(clusterInfo *model.ClusterInfo, conciliationTime time.Duration, apiVersion string) {
	i := 0
	n := len(clusterInfo.ManagerAddresses)
	for {
		log.Printf("Using manager %s", clusterInfo.ManagerAddresses[i])

		nodesUrl := fmt.Sprintf("http://%s/%s/nodes", clusterInfo.ManagerAddresses[i], apiVersion)
		nodeList, err := getNodesList(nodesUrl)
		if err == nil {
			log.Printf("Updating nodes list to: %+q", nodeList)
			clusterInfo.NodeIPs = nodeList
		} else {
			log.Printf("Cannot update nodes list due to: %s", err.Error())
		}

		servicesUrl := fmt.Sprintf("http://%s/%s/services", clusterInfo.ManagerAddresses[i], apiVersion)
		publishedPorts, err := getPublishedPorts(servicesUrl)
		if err == nil {
			log.Printf("Updating published ports to: %+q", publishedPorts)
			clusterInfo.PublishedPorts = publishedPorts
		} else {
			log.Printf("Cannot update published ports due to: %s", err.Error())
		}

		i = (i + 1) % n
		time.Sleep(conciliationTime)
	}
}

func getNodesList(url string) ([]string, error) {
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
				if node.Status.State == swarm.NodeStateReady && node.Status.Addr != "0.0.0.0" {
					nodeIPs = append(nodeIPs, node.Status.Addr)
				}
			}
		}
	} else {
		err = errors.New(fmt.Sprintf("Received http status code %d and response body: %s", resp.StatusCode, body))
	}
	return nodeIPs, err
}

func getPublishedPorts(url string) (map[string]uint32, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Received http status code %d and error while trying to read body: %v", resp.StatusCode, err))
	}

	var publishedPorts map[string]uint32
	if resp.StatusCode < 300 {
		var data []swarm.Service

		err = json.Unmarshal(body, &data)
		if err == nil {
			publishedPorts = make(map[string]uint32)
			for _, service := range data {
				if service.Endpoint.Ports != nil {
					for _, portSpec := range service.Endpoint.Ports {
						publishedPorts[service.Spec.Name] = portSpec.PublishedPort
					}
				}
			}
		}
	} else {
		err = errors.New(fmt.Sprintf("Received http status code %d and response body: %s", resp.StatusCode, body))
	}
	return publishedPorts, err
}

func handle(info *model.ClusterInfo) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error while trying to read body\n")
			return
		}
		var service model.Query
		err = json.Unmarshal(body, &service)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error while trying to parse body: %+q\n", body)
			return
		}
		publishedPort, ok := info.PublishedPorts[service.ServiceName]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Could not find published port for %s service\n", service.ServiceName)
			return
		}
		serviceInfo := model.ServiceInfo{NodeIPs:info.NodeIPs, PublishedPort: publishedPort}
		resp, err := json.Marshal(serviceInfo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Could serialize json due to: %v\n", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}
