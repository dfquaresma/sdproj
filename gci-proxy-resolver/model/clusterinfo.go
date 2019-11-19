package model

type ClusterInfo struct {
	ManagerAddresses []string
	NodeIPs []string
}

type NodeStatus struct {
	State string
	Addr string
}