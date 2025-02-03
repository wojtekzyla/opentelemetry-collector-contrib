// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package kafkatopicsobserver // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/kafkatopicsobserver"

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/kafka"
	"regexp"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
)

// Mock implementation for the initial PR
var (
	_ extension.Extension      = (*kafkaTopicsObserver)(nil)
	_ observer.EndpointsLister = (*kafkaTopicsObserver)(nil)
	_ observer.Observable      = (*kafkaTopicsObserver)(nil)
)

type kafkaTopicsObserver struct {
	*observer.EndpointsWatcher
	logger     *zap.Logger
	config     *Config
	cancel     func()
	once       *sync.Once
	ctx        context.Context
	topics     []string
	kafkaAdmin sarama.ClusterAdmin
}

func newObserver(logger *zap.Logger, config *Config) (extension.Extension, error) {
	kCtx, cancel := context.WithCancel(context.Background())

	admin, err := createKafkaClusterAdmin(kCtx, *config)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("could not create kafka cluster admin: %w", err)
	}

	d := &kafkaTopicsObserver{
		logger:     logger,
		config:     config,
		once:       &sync.Once{},
		topics:     []string{},
		ctx:        kCtx,
		cancel:     cancel,
		kafkaAdmin: admin,
	}
	d.EndpointsWatcher = observer.NewEndpointsWatcher(d, time.Second, logger)
	return d, nil
}

func (k *kafkaTopicsObserver) ListEndpoints() []observer.Endpoint {
	endpoints := make([]observer.Endpoint, 0, len(k.topics))
	for _, topic := range k.topics {
		details := &observer.KafkaTopic{
			Brokers: k.config.Brokers,
		}
		endpoint := observer.Endpoint{
			ID:      observer.EndpointID(topic),
			Target:  topic,
			Details: details,
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints
}

func (k *kafkaTopicsObserver) Start(_ context.Context, _ component.Host) error {
	k.once.Do(
		func() {
			go func() {
				topicsRefreshTicker := time.NewTicker(k.config.TopicsSyncInterval)
				defer topicsRefreshTicker.Stop()
				for {
					select {
					case <-k.ctx.Done():
						err := k.kafkaAdmin.Close()
						if err != nil {
							k.logger.Error("failed to close kafka cluster admin", zap.Error(err))
						} else {
							k.logger.Info("kafka cluster admin closed")
						}
						return
					case <-topicsRefreshTicker.C:
						// Collect all available topics
						topics, err := k.kafkaAdmin.ListTopics()
						if err != nil {
							k.logger.Error("failed to list topics: ", zap.Error(err))
							continue
						}
						var topicNames []string
						for topic := range topics {
							topicNames = append(topicNames, topic)
						}
						// Filter topics
						var subTopics []string
						reg, _ := regexp.Compile(k.config.TopicRegex)
						for _, t := range topicNames {
							if reg.MatchString(t) {
								subTopics = append(subTopics, t)
							}
						}
						// Check if there is a change in topics
						if !equalStringSlices(k.topics, subTopics) {
							k.logger.Info("change in topics detected")
							k.topics = subTopics
						}
					}
				}
			}()
		})
	return nil
}

func (k *kafkaTopicsObserver) Shutdown(_ context.Context) error {
	k.StopListAndWatch()
	k.cancel()
	return nil
}

var createKafkaClusterAdmin = func(ctx context.Context, config Config) (sarama.ClusterAdmin, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Session.Timeout = config.SessionTimeout
	saramaConfig.Consumer.Group.Heartbeat.Interval = config.HeartbeatInterval

	var err error
	if config.ResolveCanonicalBootstrapServersOnly {
		saramaConfig.Net.ResolveCanonicalBootstrapServers = true
	}
	if config.ProtocolVersion != "" {
		if saramaConfig.Version, err = sarama.ParseKafkaVersion(config.ProtocolVersion); err != nil {
			return nil, err
		}
	}
	if err := kafka.ConfigureAuthentication(ctx, config.Authentication, saramaConfig); err != nil {
		return nil, err
	}
	return sarama.NewClusterAdmin(config.Brokers, saramaConfig)
}

// equalStringSlices checks if two slices of strings have the same content
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aMap := make(map[string]struct{}, len(a))
	for _, v := range a {
		aMap[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := aMap[v]; !ok {
			return false
		}
	}
	return true
}
