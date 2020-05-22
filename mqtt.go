package goMicroServiceStat

import (
	"net/url"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTReporterConfig struct {
	ClientID  string
	BrokerURI *url.URL
	UserName  string
	Password  string
}
type MQTTReporter struct {
	config *MQTTReporterConfig
	client mqtt.Client
}

func NewMQTTReporter(config *MQTTReporterConfig) (*MQTTReporter, error) {
	m := new(MQTTReporter)
	err := m.connect(config)
	if err != nil {
		m = nil
	}
	return m, err
}

func (m *MQTTReporter) IsConnected() bool {
	if m.client == nil {
		return false
	}
	return m.client.IsConnected()
}

func (m *MQTTReporter) Disconnect() {
	if m.client != nil && m.client.IsConnected() {
		m.client.Disconnect(500)
		m.client = nil
	}
}

func (m *MQTTReporter) connect(config *MQTTReporterConfig) error {
	m.Disconnect()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.BrokerURI.String())
	if config.UserName != "" {
		opts.SetUsername(config.UserName)
		if config.Password != "" {
			opts.SetPassword(config.Password)
		}
	}
	opts.SetClientID(config.ClientID)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		return err
	}
	m.client = client
	return nil
}

func (m *MQTTReporter) Publish(topic string, qos byte, retained bool, payload []byte) {
	m.client.Publish(topic, qos, retained, payload)
}
