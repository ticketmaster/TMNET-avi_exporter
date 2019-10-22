package main

import (
	"time"

	"github.com/avinetworks/sdk/go/clients"
	"github.com/prometheus/client_golang/prometheus"
)

type guages map[string]*prometheus.GaugeVec

// Connection describes the connection.
type Connection struct {
	UserName string
	Password string
	Host     string
	Tenant   string
}

// ClusterResponse describes the response from the cluster.
type ClusterResponse struct {
	NodeInfo struct {
		MgmtIP string `json:"mgmt_ip"`
	} `json:"node_info"`
}

// CollectionResponse describes the response sent from Avi's metrics/collection endpoint.
type CollectionResponse struct {
	Header struct {
		Statistics struct {
			Min        float64   `json:"min"`
			Trend      float64   `json:"trend"`
			Max        float64   `json:"max"`
			MaxTs      time.Time `json:"max_ts"`
			MinTs      time.Time `json:"min_ts"`
			NumSamples int       `json:"num_samples"`
			Mean       float64   `json:"mean"`
		} `json:"statistics"`
		MetricsMinScale      float64 `json:"metrics_min_scale"`
		MetricDescription    string  `json:"metric_description"`
		MetricsSumAggInvalid bool    `json:"metrics_sum_agg_invalid"`
		TenantUUID           string  `json:"tenant_uuid"`
		Priority             bool    `json:"priority"`
		EntityUUID           string  `json:"entity_uuid"`
		Units                string  `json:"units"`
		ObjIDType            string  `json:"obj_id_type"`
		DerivationData       struct {
			DerivationFn          string `json:"derivation_fn"`
			SecondOrderDerivation bool   `json:"second_order_derivation"`
			MetricIds             string `json:"metric_ids"`
		} `json:"derivation_data"`
		Name string `json:"name"`
	} `json:"header"`
	Data []struct {
		Timestamp time.Time `json:"timestamp"`
		Value     float64   `json:"value"`
	} `json:"data"`
}

// MetricList is the marshalled return payload of default metrics on Avi.
type MetricList struct {
	MetricsData map[string]struct {
		EntityTypes []string `json:"entity_types"`
		MetricUnits string   `json:"metric_units"`
		Description string   `json:"description"`
	} `json:"metrics_data"`
}

// connectionOpts describes the avi connection options.
type connectionOpts struct {
	username   string
	password   string
	tenant     string
	cluster    string
	apiVersion string
}

// DefaultMetrics describes the default list of Avi metrics.
type DefaultMetrics []struct {
	Metric string `json:"metric"`
	Help   string `json:"help"`
}

// Exporter describes the prometheus exporter.
type Exporter struct {
	GaugeOptsMap     GaugeOptsMap
	AviClient        *clients.AviClient
	connectionOpts   connectionOpts
	userMetricString string
	guages           guages
}

// Gauge describes the prometheus gauge.
type Gauge struct {
	Name   string
	Entity string
	Units  string
	Value  float64
	Tenant string
	Leader string
}

// GaugeOptsMap lists all the GaugeOpts that will be registered.
type GaugeOptsMap map[string]GaugeOpts

// GaugeOpts describes the custom GaugeOpts definition for mapping.
type GaugeOpts struct {
	Type         string
	GaugeOpts    prometheus.GaugeOpts
	CustomLabels []string
}

// Metrics contains all the metrics.
type Metrics struct {
	MetricRequests []MetricRequest `json:"metric_requests"`
}

// MetricRequest describes the metric.
type MetricRequest struct {
	Step         int    `json:"step"`
	Limit        int    `json:"limit"`
	EntityUUID   string `json:"entity_uuid"`
	MetricEntity string `json:"metric_entity,omitempty"`
	MetricID     string `json:"metric_id"`
}

type virtualServiceDef struct {
	Name      string
	PoolUUID  string
	IPAddress string `json:"ipaddress"`
	FQDN      string `json:"fqdn"`
}

type clusterDef struct {
	IPAddress string `json:"ipaddress"`
	FQDN      string `json:"fqdn"`
	Name      string `json:"name"`
}

type seDef struct {
	IPAddress string `json:"ipaddress"`
	FQDN      string `json:"fqdn"`
	Name      string `json:"name"`
}

type poolDef struct {
	Name string
}

type cluster struct {
	VirtualIP struct {
		Type string `json:"type"`
		Addr string `json:"addr"`
	} `json:"virtual_ip"`
	Nodes []struct {
		IP struct {
			Type string `json:"type"`
			Addr string `json:"addr"`
		} `json:"ip"`
		VMHostname string `json:"vm_hostname"`
		VMUUID     string `json:"vm_uuid"`
		Name       string `json:"name"`
		VMMor      string `json:"vm_mor"`
	} `json:"nodes"`
	TenantUUID string `json:"tenant_uuid"`
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
}
