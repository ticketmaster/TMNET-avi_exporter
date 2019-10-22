package main

// Runtime object.
type Runtime struct {
	NodeInfo struct {
		UUID        string `json:"uuid"`
		Version     string `json:"version"`
		MgmtIP      string `json:"mgmt_ip"`
		ClusterUUID string `json:"cluster_uuid"`
	} `json:"node_info"`
	NodeStates []struct {
		MgmtIP string `json:"mgmt_ip"`
		Role   string `json:"role"`
	} `json:"node_states"`
}
