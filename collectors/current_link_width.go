package collectors

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type LinkWidthCollector struct {
	LinkWidthMetric *prometheus.Desc
}

func NewLinkWidthCollector() *LinkWidthCollector {
	return &LinkWidthCollector{
		LinkWidthMetric: prometheus.NewDesc(
			"pci_device_link_width",
			"The link width of the pci device",
			[]string{"device"},
			nil,
		),
	}
}

func (collector *LinkWidthCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.LinkWidthMetric
}

func (collector *LinkWidthCollector) Collect(wg *sync.WaitGroup, ch chan<- prometheus.Metric, slot string) {
	linkWidthFilePath := filepath.Join(PciDevicesPath, slot, "current_link_width")
	if !fileExists(linkWidthFilePath) {
		wg.Done()
		return
	}
	data, err := os.ReadFile(linkWidthFilePath)
	if err != nil {
		fmt.Printf("could not get link width for slot %s\n", slot)
		wg.Done()
		return
	}
	value, err := getFloatFromLinkWidth(strings.TrimSpace(string(data)))
	if err != nil {
		fmt.Printf("Could not parse link width from slot %s\n", slot)
		wg.Done()
		return
	}
	ch <- prometheus.MustNewConstMetric(collector.LinkWidthMetric, prometheus.GaugeValue, value, slot)
	wg.Done()
}

func getFloatFromLinkWidth(st string) (float64, error) {
	floatValue, err := strconv.ParseFloat(st, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse float value: %v\n", err)
	}

	return floatValue, nil
}
