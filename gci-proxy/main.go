package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dfquaresma/sdproj/gci-proxy-resolver/model"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/tcplisten"
)

const (
	defaultPort        = "3000"
	defaultPortUsage   = "default server port, '3000'"
	defaultTarget      = "http://127.0.0.1:8082"
	defaultTargetUsage = "default redirect url, 'http://127.0.0.1:8082'"
)

// Flags.
var (
	port       = flag.String("port", defaultPort, defaultPortUsage)
	target     = flag.String("target", defaultTarget, defaultTargetUsage)
	yGen       = flag.Int64("ygen", 0, "Young generation size, in bytes.")
	printGC    = flag.Bool("print_gc", true, "Whether to print gc information.")
	gciTarget  = flag.String("gci_target", defaultTarget, defaultTargetUsage)
	gciCmdPath = flag.String("gci_path", "", "URl path to be appended to the target to send GCI commands.")
	disableGCI = flag.Bool("disable_gci", false, "Whether to disable the GCI protocol (used to measure the raw proxy overhead")
	gateway = flag.String("gateway", "", "The OpenFaaS gateway address and port")
	serviceName = flag.String("service_name", "", "The OpenFaaS function name which GCI will be attached")
	useMesh   = flag.Bool("use_mesh", false, "To identify if must use a transport which includes routing mesh help")
)

func checkFunction(functionUpWg *sync.WaitGroup) {
	for {
		conn, err := net.Dial("tcp", *target)
		if err != nil {
			time.Sleep(time.Millisecond * 10)
			continue
		}
		conn.Close()
		log.Print("Function is up and running!")
		functionUpWg.Done()
		return
	}
}

func getServiceInfo(serviceInfo *model.ServiceInfo) {
	if *useMesh {
		url := fmt.Sprintf("http://%s/", *gateway)
		reqBody, err := json.Marshal(model.Query{ServiceName: *serviceName})
		if err != nil {
			log.Fatalf("Could not resolve service info to service %s due to %s\n", *serviceName, err.Error())
		}
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			log.Fatalf("Could not resolve service info to service %s due to %s\n", *serviceName, err.Error())
		}
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Could not resolve service info to service %s due to %s\n", *serviceName, err.Error())
		}
		var tmpSInfo *model.ServiceInfo
		err = json.Unmarshal(respBody, tmpSInfo)
		if err != nil {
			log.Fatalf("Could not deserialize json due to %s\n", err.Error())
		}
		if len(tmpSInfo.NodeIPs) == 0 {
			log.Fatal("Cannot redirect to mesh if there is no available nodes\n")
		}
		serviceInfo.NodeIPs = tmpSInfo.NodeIPs
		serviceInfo.PublishedPort = tmpSInfo.PublishedPort
	}
}

func main() {
	flag.Parse()

	if *yGen == 0 {
		log.Fatalf("ygen can not be 0. ygen:%d", *yGen)
	}
	cfg := tcplisten.Config{
		ReusePort: true,
	}
	ln, err := cfg.NewListener("tcp4", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("cannot listen to -in=%q: %s", fmt.Sprintf(":%s", *port), err)
	}
	var functionUpWg sync.WaitGroup
	serviceInfo := model.ServiceInfo{NodeIPs: make([]string, 0, 0), PublishedPort:0}
	functionUpWg.Add(1)
	var t *transport
	if *useMesh {
		t = newMeshedTransport(*target, &serviceInfo, *gciTarget, *gciCmdPath, *yGen, *printGC, &functionUpWg)
	} else {
		t = newTransport(*target, *yGen, *printGC, *gciTarget, *gciCmdPath)
	}
	s := fasthttp.Server{
		Handler:      t.RoundTrip,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go getServiceInfo(&serviceInfo)
	go checkFunction(&functionUpWg)
	if err := s.Serve(ln); err != nil {
		log.Fatalf("error in fasthttp server: %s", err)
	}
}
