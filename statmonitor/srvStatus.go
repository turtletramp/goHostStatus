package statmonitor

import (
	"log"
	"os"
	"syscall"
	"time"
)

type SrvStatConfig struct {
	RefreshInterval  time.Duration
	thresholdPercent int
}

var DefaultStatMonitorConfig *SrvStatConfig = &SrvStatConfig{
	RefreshInterval:  3 * time.Second,
	thresholdPercent: 3,
}

type SrvStat struct {
	Hostname       string
	SysTotalRamMB  int
	SysFreeRamMB   int
	SysTotalSwapMB int
	SysFreeSwapMB  int
}
type SrvStatInternal struct {
	SrvStat

	ThresholdPercent int
}

type SrvStatMonitor struct {
	config         *SrvStatConfig
	changeReceiver chan *SrvStat

	timer *time.Timer
	stop  chan struct{}
	// reportCurrentStatusNow is used to bypass the timer and force a report now
	forceReportCurrentStatus chan bool
}

func NewSrvStatMonitor(config *SrvStatConfig, changeReceiver chan *SrvStat) *SrvStatMonitor {
	mon := new(SrvStatMonitor)
	mon.config = config
	mon.changeReceiver = changeReceiver
	mon.Restart()
	return mon
}

func (mon *SrvStatMonitor) ForceReportNow() {
	mon.forceReportCurrentStatus <- true
}

// Restart will restart monitoring. (usually used if paused before)
func (mon *SrvStatMonitor) Restart() {
	mon.Stop()
	mon.monitor()
}

func abs(val int) int {
	if val < 0 {
		return -val
	}
	return val
}
func isAboveThreshold(value1 int, value2 int, thresholdPercent int, dbgText string) bool {
	if thresholdPercent < 1 {
		return true
	}
	valDiff := abs(value1 - value2)
	threshold := value1 / 100 * thresholdPercent
	if valDiff > threshold {
		log.Println(dbgText, "was above thershold")
	}
	return valDiff < threshold
}

func (stat *SrvStatInternal) Equal(stat2 *SrvStatInternal) bool {
	if stat2 == nil || stat == nil {
		return stat == stat2
	}
	return stat.Hostname == stat2.Hostname &&
		isAboveThreshold(stat.SysFreeRamMB, stat2.SysFreeRamMB, stat.ThresholdPercent, "freeRam") &&
		isAboveThreshold(stat.SysFreeSwapMB, stat2.SysFreeSwapMB, stat.ThresholdPercent, "freeSwap") &&
		isAboveThreshold(stat.SysTotalRamMB, stat2.SysTotalRamMB, stat.ThresholdPercent, "totalRam") &&
		isAboveThreshold(stat.SysTotalSwapMB, stat2.SysTotalSwapMB, stat.ThresholdPercent, "toalSwap")
}

func (mon *SrvStatMonitor) getCurrentStat() *SrvStatInternal {
	result := new(SrvStatInternal)

	// so that we can finally use it in the equal function
	result.ThresholdPercent = mon.config.thresholdPercent

	result.Hostname, _ = os.Hostname()

	in := &syscall.Sysinfo_t{}
	err := syscall.Sysinfo(in)
	if err != nil {
		log.Println("Failed to read sysinfo: " + err.Error())
	} else {
		result.SysTotalRamMB = int(in.Totalram / uint64(1024) / uint64(1024) * uint64(in.Unit))
		result.SysFreeRamMB = int(in.Freeram / uint64(1024) / uint64(1024) * uint64(in.Unit))
		result.SysTotalSwapMB = int(in.Totalswap / uint64(1024) / uint64(1024) * uint64(in.Unit))
		result.SysFreeSwapMB = int(in.Freeswap / uint64(1024) / uint64(1024) * uint64(in.Unit))
	}

	return result
}

func (mon *SrvStatMonitor) queryAndHandleNewStatus(lastStat *SrvStatInternal) *SrvStatInternal {
	newStat := mon.getCurrentStat()

	if !newStat.Equal(lastStat) {
		lastStat = newStat
		mon.changeReceiver <- &newStat.SrvStat
	}
	mon.timer.Reset(mon.config.RefreshInterval)

	return lastStat
}

func (mon *SrvStatMonitor) monitor() {
	mon.timer = time.NewTimer(mon.config.RefreshInterval)
	mon.stop = make(chan struct{})
	mon.forceReportCurrentStatus = make(chan bool)

	// start background monitoring
	go func() {
		// immediately read and report current status
		var lastStat *SrvStatInternal = mon.queryAndHandleNewStatus(nil)
		for {
			select {
			case <-mon.stop:
				return
			case <-mon.forceReportCurrentStatus:
				lastStat = mon.queryAndHandleNewStatus(nil)
			case <-mon.timer.C:
				lastStat = mon.queryAndHandleNewStatus(lastStat)
			}
		}
	}()
}

// Stop stops monitoring.
func (mon *SrvStatMonitor) Stop() {
	if mon.stop != nil {
		close(mon.stop)
		close(mon.forceReportCurrentStatus)
		mon.stop = nil
	}
}
