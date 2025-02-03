// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package kafkatopicsobserver // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/kafkatopicsobserver"

import (
	"fmt"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/kafka"

	"go.opentelemetry.io/collector/confmap"
)

// Config defines configuration for docker observer
type Config struct {
	// The list of kafka brokers (default localhost:9092)
	Brokers []string `mapstructure:"brokers"`
	// ResolveCanonicalBootstrapServersOnly makes Sarama do a DNS lookup for
	// each of the provided brokers. It will then do a PTR lookup for each
	// returned IP, and that set of names becomes the broker list. This can be
	// required in SASL environments.
	ResolveCanonicalBootstrapServersOnly bool `mapstructure:"resolve_canonical_bootstrap_servers_only"`
	// Kafka protocol version
	ProtocolVersion string `mapstructure:"protocol_version"`
	// Session interval for the Kafka consumer
	SessionTimeout time.Duration `mapstructure:"session_timeout"`
	// Heartbeat interval for the Kafka consumer
	HeartbeatInterval  time.Duration        `mapstructure:"heartbeat_interval"`
	Authentication     kafka.Authentication `mapstructure:"auth"`
	TopicRegex         string               `mapstructure:"topic_regex"`
	TopicsSyncInterval time.Duration        `mapstructure:"topics_sync_interval"`
}

func (config *Config) Validate() error {
	if len(config.Brokers) == 0 {
		return fmt.Errorf("brokers list must be specified")
	}
	if len(config.ProtocolVersion) == 0 {
		return fmt.Errorf("protocol_version must be specified")
	}
	if len(config.TopicRegex) == 0 {
		return fmt.Errorf("topic_regex must be specified")
	}
	if config.TopicsSyncInterval <= 0 {
		return fmt.Errorf("topics_sync_interval must be greater than 0")
	}
	if config.SessionTimeout <= 0 {
		return fmt.Errorf("session_timeout must be greater than 0")
	}
	if config.HeartbeatInterval <= 0 {
		return fmt.Errorf("heartbeat_interval must be greater than 0")
	}
	return nil
}

func (config *Config) Unmarshal(conf *confmap.Conf) error {
	err := conf.Unmarshal(config)
	if err != nil {
		return err
	}

	return err
}
