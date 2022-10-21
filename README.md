# spotinst-metrics-exporter

[![Build Status](https://github.com/Bonial-International-GmbH/spotinst-metrics-exporter/actions/workflows/ci.yml/badge.svg)](https://github.com/Bonial-International-GmbH/spotinst-metrics-exporter/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Bonial-International-GmbH/spotinst-metrics-exporter)](https://goreportcard.com/report/github.com/Bonial-International-GmbH/spotinst-metrics-exporter)
![License](https://img.shields.io/github/license/Bonial-International-GmbH/spotinst-metrics-exporter)

A prometheus exporter for metrics from Spotinst.

## Current features

- Ocean AWS cost metrics for ocean clusters, namespaces and workloads
- Ocean AWS resource suggestions ("right sizing") for workloads and their containers

## Building

Run `make build` to build a local binary, or `make docker-build` to build a docker image.

## Configuration

The `spotinst-metrics-exporter` requires the `SPOTINST_ACCOUNT` and
`SPOTINST_TOKEN` environment variables to be set. Furthermore you can configure
the listen address via the `--listen-address` flag.

The exporter will listen on `0.0.0.0:8080` by default and exposes prometheus
metrics at `/metrics` and a health endpoint at `/healthz`.

### Custom labels

Certain metrics also support appending Kubernetes resource labels to the metric
labels. You can control the labels that should be appended via the
`SPOTINST_CUSTOM_LABEL_NAMES` environment variable, which accepts a
comma-separated list of Kubernetes resource label names.

Example: if `SPOTINST_CUSTOM_LABEL_NAMES` contains
`app.kubernetes.io/name,team`, these labels will also be attached to the
metrics that support it.

Label names get sanitized so that they are valid prometheus label names, e.g.
`app.kubernetes.io/name` gets sanitized to `app_kubernetes_io_name`. Custom
labels that are not present on a Kubernetes resource will result in empty label
values for the respective metric labels.

Right now only the metrics `spotinst_ocean_aws_namespace_cost` and
`spotinst_ocean_aws_workload_cost` support this feature.

## Deployment

The helm chart provided in this repository can be used to deploy the metrics exporter.

First, add the helm repository:

```sh
helm repo add spotinst-metrics-exporter \
  https://bonial-international-gmbh.github.io/spotinst-metrics-exporter
```

Create a `values.yaml` and add a `spotinst` section with the account ID and
token, for example:

```yaml
---
spotinst:
  account: act-12345678
  token: the-spotinst-token
```

For more helm configuration options have a look into the [`values.yaml`
defaults](https://github.com/Bonial-International-GmbH/spotinst-metrics-exporter/blob/main/charts/spotinst-metrics-exporter/values.yaml).

Finally use helm to install the metrics exporter:

```sh
helm upgrade spotinst-metrics-exporter spotinst-metrics-exporter/spotinst-metrics-exporter \
  --install --namespace kube-system --values values.yaml
```

Alternatively, you can also pass `spotinst.account` and `spotinst.token` to the
`helm` command directly instead of using a `values.yaml` file:

```sh
helm upgrade spotinst-metrics-exporter spotinst-metrics-exporter/spotinst-metrics-exporter \
  --install --namespace kube-system \
  --set spotinst.account=act-12345678,spotinst.token=the-spotinst-token
```

## Metrics

All metrics are gauge values. The values of CPU metrics are in milli-CPU,
memory values are in MiB and cost values are in $USD. Cost metrics display the
running costs of the current month and are reset on every 1st.

### Samples

```
spotinst_ocean_aws_cluster_cost{ocean_id="o-12345678",ocean_name="my-ocean"} 301.86862
spotinst_ocean_aws_namespace_cost{namespace="kube-system",ocean_id="o-12345678",ocean_name="my-ocean"} 28.858004
spotinst_ocean_aws_workload_cost{name="coredns",namespace="kube-system",ocean_id="o-12345678",ocean_name="my-ocean",workload="deployment"} 1.2382613
spotinst_ocean_aws_workload_container_cpu_requested{container="coredns",name="coredns",namespace="kube-system",ocean_id="o-12345678",ocean_name="my-ocean",workload="deployment"} 100
spotinst_ocean_aws_workload_container_cpu_suggested{container="coredns",name="coredns",namespace="kube-system",ocean_id="o-12345678",ocean_name="my-ocean",workload="deployment"} 100
spotinst_ocean_aws_workload_container_memory_requested{container="coredns",name="coredns",namespace="kube-system",ocean_id="o-12345678",ocean_name="my-ocean",workload="deployment"} 70
spotinst_ocean_aws_workload_container_memory_suggested{container="coredns",name="coredns",namespace="kube-system",ocean_id="o-12345678",ocean_name="my-ocean",workload="deployment"} 34
spotinst_ocean_aws_workload_cpu_requested{name="coredns",namespace="kube-system",ocean_id="o-12345678",ocean_name="my-ocean",workload="deployment"} 100
spotinst_ocean_aws_workload_cpu_suggested{name="coredns",namespace="kube-system",ocean_id="o-12345678",ocean_name="my-ocean",workload="deployment"} 100
spotinst_ocean_aws_workload_memory_requested{name="coredns",namespace="kube-system",ocean_id="o-12345678",ocean_name="my-ocean",workload="deployment"} 70
spotinst_ocean_aws_workload_memory_suggested{name="coredns",namespace="kube-system",ocean_id="o-12345678",ocean_name="my-ocean",workload="deployment"} 34
```

## License

The source code of spotinst-metrics-exporter is released under the MIT License.
See the bundled LICENSE file for details.
