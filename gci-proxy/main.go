package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"
	"net"

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
	meshTarget = flag.String("mesh_target", "", "the routing mesh target to redirect instead of return 503")
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
	functionUpWg.Add(1)
	var t *transport
	if *useMesh {
		t = newMeshedTransport(*target, *meshTarget, *gciTarget, *gciCmdPath, *yGen, *printGC, &functionUpWg)
	} else {
		t = newTransport(*target, *yGen, *printGC, *gciTarget, *gciCmdPath)
	}
	s := fasthttp.Server{
		Handler:      t.RoundTrip,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go checkFunction(&functionUpWg)
	if err := s.Serve(ln); err != nil {
		log.Fatalf("error in fasthttp server: %s", err)
	}
}
