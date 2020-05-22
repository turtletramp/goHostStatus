package goMicroServiceStat

import (
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_mqtt(t *testing.T) {
	uri, err := url.Parse("tcp://localhost:1883")
	assert.NoError(t, err)

	m, err := NewMQTTReporter(&MQTTReporterConfig{
		BrokerURI: uri,
		ClientID:  "test",
	})

	assert.NoError(t, err, "failed to create MQTTReporter")
	assert.True(t, m.IsConnected())

	for i := 1; i < 50; i++ {
		m.Publish("microServiceStat/test", 2, true, []byte("hello "+strconv.Itoa(i)))
		time.Sleep(1 * time.Second)
	}

	m.Disconnect()
	assert.False(t, m.IsConnected())
}
