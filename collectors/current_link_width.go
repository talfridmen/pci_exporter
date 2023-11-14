package collectors

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type LinkWidthCollector struct {
	RegionMetric *prometheus.Desc
}

func NewLinkWidthCollector() *LinkWidthCollector {
	return &LinkWidthCollector{
		RegionMetric: prometheus.NewDesc(
			"pci_device_link_width",
			"The link width of the pci device",
			[]string{"device"},
			nil,
		),
	}
}

func (collector *LinkWidthCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.RegionMetric
}

func (collector *LinkWidthCollector) Collect(wg *sync.WaitGroup, ch chan<- prometheus.Metric, slot string) {
	linkWidthFilePath := filepath.Join(PciDevicesPath, slot, "current_link_width")
	data, err := os.ReadFile(linkWidthFilePath)
	if err != nil {
		fmt.Printf("could not get link width for slot %s\n", slot)
		return
	}
	value, err := getFloatFromLinkWidth(string(data))
	if err != nil {
		fmt.Printf("Could not parse link width from slot %s\n", slot)
		return
	}
	ch <- prometheus.MustNewConstMetric(collector.RegionMetric, prometheus.GaugeValue, value, slot)
	wg.Done()
}

func getFloatFromLinkWidth(st string) (float64, error) {
	floatValue, err := strconv.ParseFloat(st, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse float value: %v\n", err)
	}

	return floatValue, nil
}
