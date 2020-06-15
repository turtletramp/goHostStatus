package main

import (
	"log"
	"net/url"
	"os"

	"github.com/turtletramp/goMicroServiceStat/statmonitor"
)

func main() {
	conf := new(statmonitor.MQTTReporterConfig)
	conf.BrokerURI, _ = url.Parse("tcp://localhost:1883")
	conf.ClientID, _ = os.Hostname()
	conf.TopicPrefix = "monitoring/" + conf.ClientID
	conf.OnlineStatusTopic = "online"
	conf.OnlineReportPayload = []byte("true")
	conf.OfflineReportPayload = []byte("false")
	conf.OnlineStatusRetained = true

	statMon, err := statmonitor.NewSrvReport(conf)
	if err != nil {
		log.Fatal(err.Error())
	}

	forever := make(chan bool)

	log.Print("Exit with ctrl+c")

	<-forever
	statMon.Close()
}
