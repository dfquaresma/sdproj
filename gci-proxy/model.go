package main

import "sync"

type ServiceInfo struct {
	Mutex         sync.Mutex
	NodeIPs       []string
	PublishedPort uint32
}
