// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package kafkatopicsobserver

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/kafkatopicsobserver/internal/metadata"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
	"testing"
)

type MockClusterAdmin struct {
	mock.Mock
	sarama.ClusterAdmin
}

func TestCollectEndpointsDefaultConfig(t *testing.T) {
	factory := NewFactory()
	mockAdmin := &MockClusterAdmin{}
	// Override the createKafkaClusterAdmin function to return the mock admin
	originalCreateKafkaClusterAdmin := createKafkaClusterAdmin
	createKafkaClusterAdmin = func(ctx context.Context, config Config) (sarama.ClusterAdmin, error) {
		return mockAdmin, nil
	}

	ext, err := newObserver(zap.NewNop(), factory.CreateDefaultConfig().(*Config))
	require.NoError(t, err)
	require.NotNil(t, ext)

	obvs, ok := ext.(*kafkaTopicsObserver)
	require.True(t, ok)
	kEndpoints := obvs.ListEndpoints()

	want := []observer.Endpoint{}
	createKafkaClusterAdmin = originalCreateKafkaClusterAdmin
	require.Equal(t, want, kEndpoints)
}

func TestCollectEndpointsAllConfigSettings(t *testing.T) {
	extAllSettings := loadConfig(t, component.NewIDWithName(metadata.Type, "all_settings"))
	ext, err := newObserver(zap.NewNop(), extAllSettings)
	require.NoError(t, err)
	require.NotNil(t, ext)

	obvs := ext.(*kafkaTopicsObserver)

	kEndpoints := obvs.ListEndpoints()

	want := []observer.Endpoint{
		{
			ID:     "babc5a6d7af2a48e7f52e1da26047024dcf98b737e754c9c3459bb84d1e4f80c:8080",
			Target: "127.0.0.1:8080",
			Details: &observer.Container{
				Name:        "agitated_wu",
				Image:       "nginx",
				Tag:         "1.17",
				Command:     "nginx -g daemon off;",
				ContainerID: "babc5a6d7af2a48e7f52e1da26047024dcf98b737e754c9c3459bb84d1e4f80c",
				Transport:   observer.ProtocolTCP,
				Labels: map[string]string{
					"hello":      "world",
					"maintainer": "NGINX Docker Maintainers",
					"mstumpf":    "",
				},
				Port:          8080,
				AlternatePort: 80,
				Host:          "127.0.0.1",
			},
		},
	}

	require.Equal(t, want, kEndpoints)
}
