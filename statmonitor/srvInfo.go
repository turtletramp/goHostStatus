package statmonitor

import (
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
)

type SrvInfo struct {
	Timestamp  int64  `json:"timestamp"`
	Hostname   string `json:"hostname"`
	Platform   string `json:"platform"`
	Family     string `json:"family"`
	Version    string `json:"version"`
	OS         string `json:"os"`
	KernelArch string `json:"kernelArch"`

	CPU string

	Disks []*SrvDiskInfo
}

type SrvDiskInfo struct {
	Device string `json:"device"`
	Mount  string `json:"mount"`
	FsType string `json:"fsType"`

	TotalMB     uint64  `json:"totalMB"`
	FreeMB      uint64  `json:"freeMB"`
	UsedMB      uint64  `json:"usedMB"`
	UsedPercent float64 `json:"usedPercent"`
}

func GetSrvInfo() *SrvInfo {
	inf := new(SrvInfo)

	inf.Timestamp = time.Now().Unix()
	inf.Hostname, _ = os.Hostname()

	ish, _ := host.Info()
	inf.Platform = ish.Platform
	inf.Family = ish.PlatformFamily
	inf.Version = ish.PlatformVersion
	inf.OS = ish.OS
	inf.KernelArch = ish.KernelArch

	isc, _ := cpu.Info()
	if len(isc) > 0 {
		inf.CPU = isc[0].ModelName
	}

	isp, _ := disk.Partitions(false)
	for _, p := range isp {
		if strings.HasPrefix(p.Mountpoint, "/snap") {
			continue
		}
		di := new(SrvDiskInfo)
		di.Device = p.Device
		di.FsType = p.Fstype
		di.Mount = p.Mountpoint

		usage, err := disk.Usage(p.Mountpoint)
		if err == nil {
			di.UsedMB = usage.Used / 1024 / 1024
			di.UsedPercent = usage.UsedPercent
			di.FreeMB = usage.Free / 1024 / 1024
			di.TotalMB = usage.Total / 1024 / 1024
		}

		inf.Disks = append(inf.Disks, di)
	}

	return inf
}
