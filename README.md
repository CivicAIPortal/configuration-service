# Configuration Service

The configuration service provides an simple http endpoint which provides straight forward config map values as json files. It predefines no ingress so you have to declare one in the values.yaml file under ingress section.

Note: This service requires and service account in kubernetes with access to read config maps in the namespace.

# Use Case

The service provides for UIs or any other application a simple way to delivery config map content from the cluster.

# Functionality

Translates 1:1 a config map to rest api output. Example: 

Config Map:

```
data:
    example: value

```

Rest API:

``` 
{
    "example":"value"
}

```

# Mobile compatibility tests

The repository includes mobile simulation contract tests in `mobile_contract_test.go`.

These tests model multiple Android/iOS app cohorts (including old versions) and validate that:

- `/configuration` remains backward-compatible for mobile startup usage.
- `/isAlive` remains stable.
- handler behavior on upstream config map failure remains predictable (`500`).

Run:

```bash
go test ./...
```

# Observability

The service now exposes standard health and Prometheus metrics endpoints:

- `GET /health` returns JSON health status with UTC timestamp.
- `GET /isAlive` remains available for backward compatibility.
- `GET /metrics` exposes Prometheus metrics.

Custom Prometheus metrics:

- `configuration_service_http_requests_total{method,route,status}`
- `configuration_service_http_request_duration_seconds{method,route,status}`
- `configuration_service_http_in_flight_requests`

Repository observability assets:

- Prometheus scrape config: `prometheus.yml`
- Prometheus alerts: `alerts/configuration-service-alerts.yml`
- Grafana datasource provisioning: `grafana/provisioning/datasources/prometheus.yml`
- Grafana dashboard provisioning: `grafana/provisioning/dashboards/`


