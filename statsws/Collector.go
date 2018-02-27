// Collector
// Poller to retrieve system metrics at a specified interval
package statsws

import (
  "fmt"
  "time"
  "github.com/shirou/gopsutil/cpu"
  "github.com/shirou/gopsutil/mem"
  "github.com/shirou/gopsutil/net"
)

type CpuData []*CpuCoreData

type CpuCoreData struct {
  All  float64
  Busy float64
  Load uint8
}

type MemData struct {
  Available uint64
  Total     uint64
  Used      uint64
}

type NetData map[string]*IfaceData

type IfaceData struct {
  BytesRx      uint64
  BytesRxTotal uint64
  BytesTx      uint64
  BytesTxTotal uint64
  Mac          string
  Name         string
}

type Metrics struct {
  Cpu CpuData
  Mem *MemData
  Net NetData
  Ts  int64
}

type Collector struct {
  Out      chan *Metrics
  interval time.Duration
}

func NewCollector(interval int) *Collector {
  return &Collector {
    Out:      make(chan *Metrics),
    interval: time.Duration(interval) * time.Second,
  }
}

func (c *Collector) Start() {
  fmt.Println("Starting collector")
  for {
    go c.getData()
    time.Sleep(c.interval)
  }
}

func (c *Collector) getData() {
  // CPU data
  cpus, err := cpu.Times(true)
  cpuData := make(CpuData, len(cpus))

  if err == nil {
    for i, t := range cpus {
      busy := t.User + t.System + t.Nice + t.Iowait + t.Irq +
              t.Softirq + t.Steal + t.Guest + t.GuestNice + t.Stolen
      cpuData[i] = &CpuCoreData{ All: busy + t.Idle, Busy: busy }
    }
  }

  // Memory data
  m, err := mem.VirtualMemory()
  memData := &MemData{}
  if err == nil {
    memData.Available = m.Available
    memData.Total     = m.Total
    memData.Used      = m.Used
  }

  // Network interface data
  netDataByName := make(NetData, 0)

  ifaces, err := net.Interfaces()
  if err == nil {
    for _, iface := range ifaces {
      if len(iface.HardwareAddr) < 1 { continue }
      netDataByName[iface.Name] = &IfaceData{
        Mac:  iface.HardwareAddr,
        Name: iface.Name,
      }
    }
  }

  ifaceStats, err := net.IOCounters(true)
  if err == nil {
    for _, stat := range ifaceStats {
      if iface, ok := netDataByName[stat.Name]; ok {
        iface.BytesRxTotal = stat.BytesRecv
        iface.BytesTxTotal = stat.BytesSent
      }
    }
  }

  netData := make(NetData, len(netDataByName))

  // Interpolate network interface data and key by MAC address
  idx := 0
  for _, ifaceData := range netDataByName {
    netData[ifaceData.Mac] = ifaceData
    idx++
  }

  metrics := &Metrics{
    Cpu: cpuData,
    Mem: memData,
    Net: netData,
    Ts:  time.Now().Unix(),
  }

  // Send metrics to output channel
  c.Out <- metrics
}
