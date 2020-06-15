package statmonitor

import (
	"net/url"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/labstack/gommon/log"
)

type MQTTReporterConfig struct {
	ClientID    string
	BrokerURI   *url.URL
	UserName    string
	Password    string
	TopicPrefix string

	OnlineStatusTopic    string
	OfflineReportPayload []byte
	OnlineReportPayload  []byte
	OnlineStatusRetained bool

	OnConnectionStatusChanged func(isConnected bool)
}
type MQTTReporter struct {
	config *MQTTReporterConfig
	client mqtt.Client
}

func NewMQTTReporter(config *MQTTReporterConfig) (*MQTTReporter, error) {
	m := new(MQTTReporter)
	m.config = config
	err := m.connect(config)

	if err != nil {
		m = nil
	}

	return m, err
}

// createTopicPath will add the topic prefix if configured
func (m *MQTTReporter) createTopicPath(topic string) string {
	if m.config.TopicPrefix != "" {
		return m.config.TopicPrefix + "/" + topic
	}
	return topic
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

	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		log.Warn("connection lost; will try to reconnect; err=" + err.Error())
		if m.config.OnConnectionStatusChanged != nil {
			m.config.OnConnectionStatusChanged(false)
		}
	})
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		log.Info("connection established")
		if m.config.OnConnectionStatusChanged != nil {
			m.config.OnConnectionStatusChanged(true)
		}
	})

	// is online status report wanted?
	if config.OnlineStatusTopic != "" && len(config.OfflineReportPayload) > 0 {
		// YES --> set last will
		opts.SetBinaryWill(m.createTopicPath(config.OnlineStatusTopic), config.OfflineReportPayload, 2, config.OnlineStatusRetained)
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		return err
	}

	m.client = client

	// is online status report wanted?
	if config.OnlineStatusTopic != "" && len(config.OnlineReportPayload) > 0 {
		m.client.Publish(m.createTopicPath(config.OnlineStatusTopic), 1, false, config.OnlineReportPayload)
	}

	return nil
}

func (m *MQTTReporter) Publish(topic string, qos byte, retained bool, payload []byte) mqtt.Token {
	return m.client.Publish(m.createTopicPath(topic), qos, retained, payload)
}
