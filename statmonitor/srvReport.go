package statmonitor

import (
	"encoding/json"
)

type SrvReport struct {
	mqtt            *MQTTReporter
	srvMonitor      *SrvStatMonitor
	srvStatReceiver chan *SrvStat

	doClose chan struct{}
}

func NewSrvReport(mqttConfig *MQTTReporterConfig) (*SrvReport, error) {
	sr := new(SrvReport)

	var err error
	sr.mqtt, err = NewMQTTReporter(mqttConfig)
	if err != nil {
		return nil, err
	}

	origConnectCb := mqttConfig.OnConnectionStatusChanged
	numConnections := 0
	mqttConfig.OnConnectionStatusChanged = func(isConnected bool) {
		if isConnected {
			numConnections++
			if sr.srvMonitor != nil && numConnections > 1 {
				// let's force status update on every later connection (usually happens when the broker restarts)
				sr.srvMonitor.ForceReportNow()
			}
		}
		if origConnectCb != nil {
			origConnectCb(isConnected)
		}
	}

	sr.srvStatReceiver = make(chan *SrvStat)
	sr.srvMonitor = NewSrvStatMonitor(DefaultStatMonitorConfig, sr.srvStatReceiver)

	sr.publishSrvInfo()
	sr.startPublishSrvStat()

	return sr, nil
}

func (sr *SrvReport) publishSrvInfo() {
	data, _ := json.Marshal(GetSrvInfo())
	sr.mqtt.Publish("info", 1, true, data)
}

func (sr *SrvReport) publishStat(stat *SrvStat) {
	data, _ := json.Marshal(stat)
	sr.mqtt.Publish("status", 1, false, data)
}

func (sr *SrvReport) startPublishSrvStat() {
	go func() {
		for {
			select {
			case <-sr.doClose:
				return
			case stat := <-sr.srvStatReceiver:
				sr.publishStat(stat)
			}
		}
	}()
}

func (sr *SrvReport) Close() {
	sr.srvMonitor.Stop()
	close(sr.doClose)
	sr.mqtt.Disconnect()
}
