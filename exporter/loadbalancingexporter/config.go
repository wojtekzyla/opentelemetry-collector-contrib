// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package loadbalancingexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/loadbalancingexporter"

import (
	"fmt"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
)

type routingKey int

const (
	traceIDRouting routingKey = iota
	svcRouting
	metricNameRouting
	resourceRouting
	streamIDRouting
)

const (
	svcRoutingStr        = "service"
	traceIDRoutingStr    = "traceID"
	metricNameRoutingStr = "metric"
	resourceRoutingStr   = "resource"
	streamIDRoutingStr   = "streamID"
)

// Config defines configuration for the exporter.
type Config struct {
	TimeoutSettings           exporterhelper.TimeoutConfig `mapstructure:",squash"`
	configretry.BackOffConfig `mapstructure:"retry_on_failure"`
	QueueSettings             exporterhelper.QueueConfig `mapstructure:"sending_queue"`

	Exporter   Exporter         `mapstructure:"exporter"`
	Resolver   ResolverSettings `mapstructure:"resolver"`
	RoutingKey string           `mapstructure:"routing_key"`
}

type OtlpExporter struct {
	Enabled  bool                `mapstructure:"enabled"`
	Settings otlpexporter.Config `mapstructure:"settings"`
}

type SplunkExporter struct {
	Enabled  bool                     `mapstructure:"enabled"`
	Settings splunkhecexporter.Config `mapstructure:"settings"`
}

// Protocol holds the individual protocol-specific settings.
type Exporter struct {
	OTLP   OtlpExporter   `mapstructure:"otlp"`
	Splunk SplunkExporter `mapstructure:"splunk"`
}

// ResolverSettings defines the configurations for the backend resolver
type ResolverSettings struct {
	Static      *StaticResolver      `mapstructure:"static"`
	DNS         *DNSResolver         `mapstructure:"dns"`
	K8sSvc      *K8sSvcResolver      `mapstructure:"k8s"`
	AWSCloudMap *AWSCloudMapResolver `mapstructure:"aws_cloud_map"`
}

// StaticResolver defines the configuration for the resolver providing a fixed list of backends
type StaticResolver struct {
	Hostnames []string `mapstructure:"hostnames"`
}

// DNSResolver defines the configuration for the DNS resolver
type DNSResolver struct {
	Hostname string        `mapstructure:"hostname"`
	Port     string        `mapstructure:"port"`
	Interval time.Duration `mapstructure:"interval"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

// K8sSvcResolver defines the configuration for the DNS resolver
type K8sSvcResolver struct {
	Service         string        `mapstructure:"service"`
	Ports           []int32       `mapstructure:"ports"`
	Timeout         time.Duration `mapstructure:"timeout"`
	ReturnHostnames bool          `mapstructure:"return_hostnames"`
}

type AWSCloudMapResolver struct {
	NamespaceName string                   `mapstructure:"namespace"`
	ServiceName   string                   `mapstructure:"service_name"`
	HealthStatus  types.HealthStatusFilter `mapstructure:"health_status"`
	Interval      time.Duration            `mapstructure:"interval"`
	Timeout       time.Duration            `mapstructure:"timeout"`
	Port          *uint16                  `mapstructure:"port"`
}

func (c *Config) Validate() error {
	count := 0
	if c.Exporter.OTLP.Enabled {
		count++
	}

	if c.Exporter.Splunk.Enabled {
		count++
	}

	if count > 1 {
		return fmt.Errorf("only one exporter should be specified")
	}
	return nil
}
