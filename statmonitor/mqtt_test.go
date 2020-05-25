package statmonitor

import (
	"net/url"
	"strconv"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
	"github.com/teris-io/shortid"
)

func Test_mqtt(t *testing.T) {
	uri, err := url.Parse("tcp://localhost:1883")
	assert.NoError(t, err)

	clientID, _ := shortid.Generate()
	m, err := NewMQTTReporter(&MQTTReporterConfig{
		BrokerURI: uri,
		ClientID:  "test" + clientID,
	})

	assert.NoError(t, err, "failed to create MQTTReporter")
	assert.True(t, m.IsConnected())

	topic := "microServiceStat/test"

	// let's first clean the messages (to be sure the topic does not have messages)
	// not necessary:
	//m.Publish(topic, 0, true, nil).Wait()

	// subscribe to the topic to receive messages
	recvMsgs := make(chan mqtt.Message, 10)
	m.client.Subscribe(topic, 0, func(c mqtt.Client, m mqtt.Message) {
		log.Info("received mqtt msg for topic " + m.Topic() + " with id " + strconv.Itoa(int(m.MessageID())) + "; payload=" + string(m.Payload()))
		recvMsgs <- m
	})

	// now let's send one
	payload := "hello " + time.Now().String()
	m.Publish(topic, 0, false, []byte(payload)).Wait()

	// we have to iterate the messages because we might have received multiple (from other sources)
	found := false
	for msg := range recvMsgs {
		if string(msg.Payload()) == payload {
			found = true
			break
		}
	}
	assert.True(t, found)

	m.Disconnect()
	assert.False(t, m.IsConnected())
}
