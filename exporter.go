package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/avinetworks/sdk/go/clients"
	"github.com/avinetworks/sdk/go/session"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/pretty"
)

func formatAviRef(in string) string {
	uriArr := strings.SplitAfter(in, "/")
	return uriArr[len(uriArr)-1]
}

func fromJSONFile(path string, ob interface{}) (err error) {
	toReturn := ob
	openedFile, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return
	}
	byteValue, err := ioutil.ReadAll(openedFile)
	if err != nil {
		log.Println(err)
		return
	}
	err = json.Unmarshal(byteValue, &toReturn)
	if err != nil {
		log.Println(err)
		return
	}
	defer openedFile.Close()
	return
}

func (o *Exporter) getDefaultMetrics(entityType string) (r DefaultMetrics, err error) {
	var path string
	r = DefaultMetrics{}
	switch entityType {
	case "virtualservice":
		path = "lib/virtualservice_metrics.json"
	case "serviceengine":
		path = "lib/serviceengine_metrics.json"
	case "controller":
		path = "lib/controller_metrics.json"
	default:
		err = errors.New("entity type must be either: virtualserver, servicengine or controller")
		log.Panic(err)
	}

	err = fromJSONFile(path, &r)
	if err != nil {
		log.Panic(err)
	}
	return
}

func (o *Exporter) setAllMetricsMap() (r GaugeOptsMap) {
	r = make(GaugeOptsMap)
	//////////////////////////////////////////////////////////////////////////////
	// Get default metrics.
	//////////////////////////////////////////////////////////////////////////////
	vsDefaultMetrics, err := o.getDefaultMetrics("virtualservice")
	if err != nil {
		log.Panic(err)
	}
	seDefaultMetrics, err := o.getDefaultMetrics("serviceengine")
	if err != nil {
		log.Panic(err)
	}
	controllerDefaultMetrics, err := o.getDefaultMetrics("controller")
	if err != nil {
		log.Panic(err)
	}
	//////////////////////////////////////////////////////////////////////////////
	// Populating default metrics. Leaving these as separate functions
	// in the event we want different GaugeOpts in the future.
	//////////////////////////////////////////////////////////////////////////////
	for _, v := range vsDefaultMetrics {
		fName := strings.ReplaceAll(v.Metric, ".", "_")
		r[v.Metric] = GaugeOpts{CustomLabels: []string{"name", "fqdn", "ipaddress", "pool", "tenant_uuid", "units", "cluster"}, Type: "virtualservice", GaugeOpts: prometheus.GaugeOpts{Name: fName, Help: v.Help}}
	}
	for _, v := range seDefaultMetrics {
		fName := strings.ReplaceAll(v.Metric, ".", "_")
		r[v.Metric] = GaugeOpts{CustomLabels: []string{"name", "entity_uuid", "fqdn", "ipaddress", "tenant_uuid", "units", "cluster"}, Type: "serviceengine", GaugeOpts: prometheus.GaugeOpts{Name: fName, Help: v.Help}}
	}
	for _, v := range controllerDefaultMetrics {
		fName := strings.ReplaceAll(v.Metric, ".", "_")
		r[v.Metric] = GaugeOpts{CustomLabels: []string{"name", "entity_uuid", "fqdn", "ipaddress", "tenant_uuid", "units", "cluster"}, Type: "controller", GaugeOpts: prometheus.GaugeOpts{Name: fName, Help: v.Help}}
	}
	//////////////////////////////////////////////////////////////////////////////
	return
}

func (o *Exporter) setPromMetricsMap() (r GaugeOptsMap) {
	r = make(GaugeOptsMap)
	all := o.setAllMetricsMap()
	if o.userMetricString == "" {
		r = all
		return
	}
	/////////////////////////////////////////////////////////
	// User provided metrics list
	/////////////////////////////////////////////////////////
	metrics := strings.Split(o.userMetricString, ",")
	for _, v := range metrics {
		r[v] = all[v]
	}
	return
}
func (o *Exporter) setUserMetrics() (r string) {
	r = os.Getenv("AVI_METRICS")
	return
}

// NewExporter constructor.
func NewExporter() (r *Exporter) {
	r = new(Exporter)
	r.userMetricString = r.setUserMetrics()
	r.connectionOpts = r.setConnectionOpts()
	r.GaugeOptsMap = r.setPromMetricsMap()
	return
}

func (o *Exporter) setConnectionOpts() (r connectionOpts) {
	r.username = os.Getenv("AVI_USERNAME")
	r.password = os.Getenv("AVI_PASSWORD")
	r.cluster = os.Getenv("AVI_CLUSTER")
	r.tenant = os.Getenv("AVI_TENANT")
	r.apiVersion = os.Getenv("AVI_APIVERSION")
	return
}

// connect establishes the avi connection.
func (o *Exporter) connect() (r *clients.AviClient, err error) {
	// simplify avi connection
	r, err = clients.NewAviClient(o.connectionOpts.cluster, o.connectionOpts.username,
		session.SetPassword(o.connectionOpts.password),
		session.SetTenant(o.connectionOpts.tenant),
		session.SetInsecure,
		session.SetVersion(o.connectionOpts.apiVersion))
	return
}
func (o *Exporter) registerGauges() {
	o.guages = make(map[string]*prometheus.GaugeVec)
	for k, v := range o.GaugeOptsMap {
		g := prometheus.NewGaugeVec(v.GaugeOpts, v.CustomLabels)
		prometheus.MustRegister(g)
		o.guages[k] = g
	}
}

// sortUniqueKeys sorts unique keys within a string array.
func sortUniqueKeys(in []string) ([]string, error) {
	var err error
	var resp []string
	respMap := make(map[string]string)
	for _, v := range in {
		respMap[v] = v
	}
	for _, v := range respMap {
		resp = append(resp, v)
	}
	sort.Strings(resp)
	return resp, err
}

func (o *Exporter) getVirtualServices() (r map[string]virtualServiceDef, err error) {
	vs, err := o.AviClient.VirtualService.GetAll()
	var pooluuid string

	if err != nil {
		log.Panic(err)
	}
	r = make(map[string]virtualServiceDef)
	for _, v := range vs {
		vip := v.Vip[0]
		address := *vip.IPAddress.Addr
		dns, _ := net.LookupAddr(address)
		for k, v := range dns {
			dns[k] = strings.TrimSuffix(v, ".")
		}

		dns, err = sortUniqueKeys(dns)

		if v.PoolRef != nil {
			pooluuid = formatAviRef(*v.PoolRef)
		}

		r[*v.UUID] = virtualServiceDef{Name: *v.Name, IPAddress: address, FQDN: strings.Join(dns, ","), PoolUUID: pooluuid}
	}
	return
}

func (o *Exporter) getClusterRuntime() (r map[string]clusterDef, err error) {
	resp := new(cluster)
	err = o.AviClient.AviSession.Get("/api/cluster", &resp)

	if err != nil {
		log.Panic(err)
	}
	r = make(map[string]clusterDef)
	for _, v := range resp.Nodes {
		address := v.IP.Addr
		dns, _ := net.LookupAddr(address)
		r[v.VMUUID] = clusterDef{Name: v.Name, IPAddress: address, FQDN: strings.Join(dns, ",")}
	}
	return
}

func (o *Exporter) getServiceEngines() (r map[string]seDef, err error) {
	se, err := o.AviClient.ServiceEngine.GetAll()
	if err != nil {
		log.Panic(err)
	}
	r = make(map[string]seDef)
	for _, v := range se {
		address := *v.MgmtVnic.VnicNetworks[0].IP.IPAddr.Addr
		dns, _ := net.LookupAddr(address)
		for k, v := range dns {
			dns[k] = strings.TrimSuffix(v, ".")
		}
		r[*v.UUID] = seDef{Name: *v.Name, IPAddress: address, FQDN: strings.Join(dns, ",")}
	}
	return
}

func (o *Exporter) getPools() (r map[string]poolDef, err error) {
	vs, err := o.AviClient.Pool.GetAll()
	if err != nil {
		log.Panic(err)
	}
	r = make(map[string]poolDef)
	for _, v := range vs {
		r[*v.UUID] = poolDef{Name: *v.Name}
	}
	return
}

// toPrettyJSON formats json output.
func toPrettyJSON(p interface{}) []byte {
	bytes, err := json.Marshal(p)
	if err != nil {
		log.Println(err.Error())
	}
	return pretty.Pretty(bytes)
}

// Collect retrieves metrics for Avi.
func (o *Exporter) Collect() (err error) {
	log.Println("polling")
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Connect to the cluster.
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	o.AviClient, err = o.connect()
	if err != nil {
		log.Panic(err)
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Set promMetrics.
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	err = o.setVirtualServiceMetrics()
	if err != nil {
		log.Print(err)
		return
	}
	err = o.setServiceEngineMetrics()
	if err != nil {
		log.Print(err)
		return
	}
	err = o.setControllerMetrics()
	if err != nil {
		log.Print(err)
		return
	}
	return
}

func (o *Exporter) getVirtualServiceMetrics() (r [][]CollectionResponse, err error) {
	req := Metrics{}
	for k, v := range o.GaugeOptsMap {
		if v.Type == "virtualservice" {
			reqMetric := MetricRequest{}
			reqMetric.EntityUUID = "*"
			reqMetric.MetricEntity = "VSERVER_METRICS_ENTITY"
			reqMetric.Limit = 1
			reqMetric.MetricID = k
			reqMetric.Step = 5
			req.MetricRequests = append(req.MetricRequests, reqMetric)
		}
	}

	resp := make(map[string]map[string][]CollectionResponse)
	err = o.AviClient.AviSession.Post("/api/analytics/metrics/collection", req, &resp)
	if err != nil {
		log.Panic(err)
		return
	}
	for _, s := range resp["series"] {
		r = append(r, s)
	}

	return
}

func (o *Exporter) getServiceEngineMetrics() (r [][]CollectionResponse, err error) {
	req := Metrics{}
	for k, v := range o.GaugeOptsMap {
		if v.Type == "serviceengine" {
			reqMetric := MetricRequest{}
			reqMetric.EntityUUID = "*"
			reqMetric.MetricEntity = "SE_METRICS_ENTITY"
			reqMetric.Limit = 1
			reqMetric.MetricID = k
			reqMetric.Step = 5
			req.MetricRequests = append(req.MetricRequests, reqMetric)
		}
	}

	resp := make(map[string]map[string][]CollectionResponse)
	err = o.AviClient.AviSession.Post("/api/analytics/metrics/collection", req, &resp)
	if err != nil {
		log.Panic(err)
		return
	}
	for _, s := range resp["series"] {
		r = append(r, s)
	}

	return
}

func (o *Exporter) getControllerMetrics() (r [][]CollectionResponse, err error) {
	req := Metrics{}
	for k, v := range o.GaugeOptsMap {
		if v.Type == "controller" {
			reqMetric := MetricRequest{}
			reqMetric.EntityUUID = "*"
			reqMetric.MetricEntity = "CONTROLLER_METRICS_ENTITY"
			reqMetric.Limit = 1
			reqMetric.MetricID = k
			reqMetric.Step = 5
			req.MetricRequests = append(req.MetricRequests, reqMetric)
		}
	}

	resp := make(map[string]map[string][]CollectionResponse)
	err = o.AviClient.AviSession.Post("/api/analytics/metrics/collection", req, &resp)
	if err != nil {
		log.Panic(err)
		return
	}
	for _, s := range resp["series"] {
		r = append(r, s)
	}

	return
}

func (o *Exporter) setVirtualServiceMetrics() (err error) {
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Get lb objects for mapping.
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	vs, _ := o.getVirtualServices()
	pools, _ := o.getPools()
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	results, err := o.getVirtualServiceMetrics()
	if err != nil {
		log.Panic(err)
		return
	}
	for _, v := range results {
		for _, v1 := range v {
			var labels prometheus.Labels
			labels = make(map[string]string)
			labels["name"] = vs[v1.Header.EntityUUID].Name
			labels["pool"] = pools[vs[v1.Header.EntityUUID].PoolUUID].Name
			labels["tenant_uuid"] = v1.Header.TenantUUID
			labels["cluster"] = o.connectionOpts.cluster
			labels["units"] = v1.Header.Units
			labels["fqdn"] = vs[v1.Header.EntityUUID].FQDN
			labels["ipaddress"] = vs[v1.Header.EntityUUID].IPAddress
			o.guages[v1.Header.Name].With(labels).Set(v1.Data[len(v1.Data)-1].Value)
		}
	}
	return
}

func (o *Exporter) setServiceEngineMetrics() (err error) {
	results, err := o.getServiceEngineMetrics()
	ses, _ := o.getServiceEngines()
	if err != nil {
		log.Panic(err)
		return
	}
	for _, v := range results {
		for _, v1 := range v {
			var labels prometheus.Labels
			labels = make(map[string]string)
			labels["tenant_uuid"] = v1.Header.TenantUUID
			labels["entity_uuid"] = v1.Header.EntityUUID
			labels["cluster"] = o.connectionOpts.cluster
			labels["units"] = v1.Header.Units
			labels["name"] = ses[v1.Header.EntityUUID].Name
			labels["fqdn"] = ses[v1.Header.EntityUUID].FQDN
			labels["ipaddress"] = ses[v1.Header.EntityUUID].IPAddress
			o.guages[v1.Header.Name].With(labels).Set(v1.Data[len(v1.Data)-1].Value)
		}
	}
	return
}

func (o *Exporter) setControllerMetrics() (err error) {
	results, err := o.getControllerMetrics()
	runtime, _ := o.getClusterRuntime()

	if err != nil {
		log.Panic(err)
		return
	}
	for _, v := range results {
		for _, v1 := range v {
			var labels prometheus.Labels
			labels = make(map[string]string)
			labels["tenant_uuid"] = v1.Header.TenantUUID
			labels["entity_uuid"] = v1.Header.EntityUUID
			labels["cluster"] = o.connectionOpts.cluster
			labels["units"] = v1.Header.Units
			labels["name"] = runtime[v1.Header.EntityUUID].Name
			labels["fqdn"] = runtime[v1.Header.EntityUUID].FQDN
			labels["ipaddress"] = runtime[v1.Header.EntityUUID].IPAddress
			o.guages[v1.Header.Name].With(labels).Set(v1.Data[len(v1.Data)-1].Value)
		}
	}
	return
}
