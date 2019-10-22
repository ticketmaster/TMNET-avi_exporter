# TMNET-AVI_EXPORTER
## Description
This is a simple Prometheus exporter that collects metrics from the Avi clusters. The exporter is written in Go and is meant to run from Kubernetes.

## Environmental Variables
| Name | Type | Description |
| ---- | ---- | ----------- |
| AVI_METRICS | string (Comma-Separated) | List of all the metrics you wish to collect. Not setting this variable defaults to ALL metrics. |
| AVI_USERNAME | string | Username for Avi Cluster |
| AVI_PASSWORD | string | Password for Avi Cluster |
| AVI_CLUSTER | string | Name of Avi Cluster (e.g., lbc.noprod1.phx.netops.tmcs) |
| AVI_TENANT | string | Name of tenant on Avi Cluster. Use 'admin' if you wish to collect all reosurces. |
| AVI_APIVERSION | string | Version running on Avi Cluster |

## Metric Files
All metric definitions are located under the `lib` directory. Each file is in JSON format, and you can update the descriptions accordingly. Feel free to use configmaps in-place of these files.

The metric files include:
- controller_metrics.json
- serviceengine_metrics.json
- virtualservice_metrics.json

In a future release, we plan to derive these metrics directly from the cluster `/api/analytics/metrics-option`; however, using a flat-file allows us to further customize the `help` attribute of the metrics.

## How it Works
Build the Docker image, using the project's Dockerfile or compile the Go binary. Before running the binary or docker image, be sure to set the environmental variables. The only variable that allows an empty value is AVI_METRICS.

During runtime, the exporter will compare the user-defined AVI_METRICS variable with the metrics listed inside of the `lib` directory. It will either match the metrics 1:1 or use all the metrics defined in the JSON files. Once the metric list is compiled, the exporter will register all the gauges and set the current value of the gauges.

Each time a GET query calls `<exporter_location>:8080/metrics`, a custom handler will invoke a collect method that will update all the registered gauges.
