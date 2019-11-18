package model

type ClusterInfo struct {
	ManagerAddresses []string
	NodeIPs []string
	PublishedPorts map[string]uint32
}

type ServiceInfo struct {
	NodeIPs []string
	PublishedPort uint32
}

type NodeStatus struct {
	State string
	Addr string
}