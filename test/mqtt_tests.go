// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	mqttModuleName         = "mqtt"
	mqttDefaultWaitTimeout = 30 * time.Second
	mqttBrokerStartupDelay = 3 * time.Second
)

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("mqtt_basic-test", mqttModuleName, "2.6.4", "", "1. 18", "", TestMQTTBasic),
		NewGeneralTestCase("mqtt_publish-test", mqttModuleName, "2.6.4", "", "1.18", "", TestMQTTPublish),
		NewGeneralTestCase("mqtt_subscribe-test", mqttModuleName, "2.6.4", "", "1.18", "", TestMQTTSubscribe),
		NewGeneralTestCase("mqtt_propagation-test", mqttModuleName, "2.6.4", "", "1.18", "", TestMQTTPropagation),
	)
}

// TestMQTTBasic tests basic MQTT functionality
func TestMQTTBasic(t *testing.T, env ...string) {
	container := initMQTTContainer(t)
	defer container.Cleanup(context.Background())

	UseApp("mqtt/v2.6.4")
	RunGoBuild(t, "go", "build", "test_mqtt_basic.go", "base. go")

	env = append(env,
		"MQTT_BROKER_ADDR="+container.BrokerAddr,
	)
	RunApp(t, "test_mqtt_basic", env...)
}

// TestMQTTPublish tests MQTT publish functionality
func TestMQTTPublish(t *testing.T, env ...string) {
	container := initMQTTContainer(t)
	defer container.Cleanup(context.Background())

	UseApp("mqtt/v2.6.4")
	RunGoBuild(t, "go", "build", "test_mqtt_publish.go", "base.go")

	env = append(env,
		"MQTT_BROKER_ADDR="+container.BrokerAddr,
	)
	RunApp(t, "test_mqtt_publish", env...)
}

// TestMQTTSubscribe tests MQTT subscribe functionality
func TestMQTTSubscribe(t *testing.T, env ...string) {
	container := initMQTTContainer(t)
	defer container.Cleanup(context.Background())

	UseApp("mqtt/v2.6.4")
	RunGoBuild(t, "go", "build", "test_mqtt_subscribe.go", "base.go")

	env = append(env,
		"MQTT_BROKER_ADDR="+container.BrokerAddr,
	)
	RunApp(t, "test_mqtt_subscribe", env...)
}

// TestMQTTPropagation tests MQTT trace propagation
func TestMQTTPropagation(t *testing.T, env ...string) {
	container := initMQTTContainer(t)
	defer container.Cleanup(context.Background())

	UseApp("mqtt/v2.6.4")
	RunGoBuild(t, "go", "build", "test_mqtt_propagation.go", "base.go")

	env = append(env,
		"MQTT_BROKER_ADDR="+container.BrokerAddr,
	)
	RunApp(t, "test_mqtt_propagation", env...)
}

// MQTTContainer holds references to the MQTT broker container
type MQTTContainer struct {
	BrokerContainer testcontainers.Container
	BrokerAddr      string
}

// Cleanup terminates the MQTT container
func (m *MQTTContainer) Cleanup(ctx context.Context) {
	if m.BrokerContainer != nil {
		_ = m.BrokerContainer.Terminate(ctx)
	}
}

// initMQTTContainer initializes the MQTT test environment with Eclipse Mosquitto broker
func initMQTTContainer(t *testing.T) *MQTTContainer {
	ctx := context.Background()

	// Start MQTT Broker (Eclipse Mosquitto) container
	brokerReq := testcontainers.ContainerRequest{
		Image:        "eclipse-mosquitto: 2.0.18",
		ExposedPorts: []string{"1883/tcp"},
		WaitingFor:   wait.ForLog("mosquitto version").WithStartupTimeout(mqttDefaultWaitTimeout),
	}

	brokerC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: brokerReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start MQTT Broker container: %v", err)
	}

	// Get the broker host and port
	host, err := brokerC.Host(ctx)
	if err != nil {
		brokerC.Terminate(ctx)
		t.Fatalf("Failed to get broker host: %v", err)
	}

	mappedPort, err := brokerC.MappedPort(ctx, "1883")
	if err != nil {
		brokerC.Terminate(ctx)
		t.Fatalf("Failed to get mapped port: %v", err)
	}

	brokerAddr := host + ":" + mappedPort.Port()

	// Wait for broker to fully initialize
	time.Sleep(mqttBrokerStartupDelay)

	return &MQTTContainer{
		BrokerContainer: brokerC,
		BrokerAddr:      brokerAddr,
	}
}
