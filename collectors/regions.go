package collectors

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type RegionCollector struct {
	RegionMetric *prometheus.Desc
}

func NewRegionCollector() *RegionCollector {
	return &RegionCollector{
		RegionMetric: prometheus.NewDesc(
			"pci_device_region_size_bytes",
			"The size of each memory region of the pci device",
			[]string{"device", "region"},
			nil,
		),
	}
}

func (collector *RegionCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.RegionMetric
}

func (collector *RegionCollector) Collect(ch chan<- prometheus.Metric, slot string) {
	slotPath := filepath.Join(PciDevicesPath, slot)

	fileList, err := os.ReadDir(slotPath)
	if err != nil {
		fmt.Printf("Error reading pci slot directory for slot %s\n", slot)
		return
	}
	for _, file := range fileList {
		if strings.HasPrefix(file.Name(), "resource") && file.Name() != "resource" {
			fileInfo, err := file.Info()
			if err != nil {
				fmt.Printf("Could not collect size for region %s in slot %s\n", file.Name(), slot)
				return
			}
			ch <- prometheus.MustNewConstMetric(collector.RegionMetric, prometheus.GaugeValue, float64(fileInfo.Size()), slot, file.Name())
		}
	}
}
