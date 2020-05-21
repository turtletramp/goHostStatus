package goMicroServiceStat

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_srvInfo(t *testing.T) {

	// setup configuration
	//srvConfig := new(SrvStatConfig)
	//piConfig.RefreshInterval: time.Second *30
	//piConfig.changeThresholdPercent: 2

	// setup channel to receive process info changes
	statChanges := make(chan (*SrvStat))

	// initialize new process info
	mon := NewSrvStatMonitor(DefaultStatMonitorConfig, statChanges)
	timeout := time.NewTimer(5 * time.Second)
	var lastStat *SrvStat
	for {
		select {
		case <-timeout.C:
			assert.NotNil(t, lastStat, "no status received")
			mon.Stop()
			return
		case stat := <-statChanges:
			lastStat = stat
			fmt.Printf("stat: %+v\n", stat)
		}
	}

}
