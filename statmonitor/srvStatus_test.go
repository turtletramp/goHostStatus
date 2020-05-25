package statmonitor

import (
	"fmt"
	"testing"
	"time"
)

func Test_srvInfo(t *testing.T) {

	// setup configuration
	//srvConfig := new(SrvStatConfig)
	//piConfig.RefreshInterval: time.Second *30
	//piConfig.changeThresholdPercent: 2

	// setup channel to receive process info changes
	statChanges := make(chan (*SrvStat))

	// initialize new process info
	config := *DefaultStatMonitorConfig
	config.RefreshInterval = 1 * time.Second
	mon := NewSrvStatMonitor(&config, statChanges)

	firstStat := <-statChanges
	fmt.Printf("stat: %+v\n", firstStat)

	// get a lot ot memory so that we should get another status info
	x := make([]byte, 1024*1024*1024*1)
	for i := range x {
		x[i] = byte(i % 256)
	}

	secondStat := <-statChanges
	fmt.Printf("stat: %+v\n", secondStat)

	mon.Stop()
}
