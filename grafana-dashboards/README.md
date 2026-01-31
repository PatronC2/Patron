# These dashboards are meant to be used with a Prometheus scraper

## Prometheus scraping config for API

```
global:
  scrape_interval: 15s
  evaluation_interval: 15s
scrape_configs:
  - job_name: "patron-api"
    scheme: https
    metrics_path: /metrics
    static_configs:
      - targets:
        - '<some-ip>:8443'

    tls_config:
      insecure_skip_verify: true
```
