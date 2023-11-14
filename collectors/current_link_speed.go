package collectors

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type LinkSpeedCollector struct {
	RegionMetric *prometheus.Desc
}

func NewLinkSpeedCollector() *LinkSpeedCollector {
	return &LinkSpeedCollector{
		RegionMetric: prometheus.NewDesc(
			"pci_device_link_speed_GTs",
			"The link speed of the pci device",
			[]string{"device"},
			nil,
		),
	}
}

func (collector *LinkSpeedCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.RegionMetric
}

func (collector *LinkSpeedCollector) Collect(wg *sync.WaitGroup, ch chan<- prometheus.Metric, slot string) {
	linkSpeedFilePath := filepath.Join(PciDevicesPath, slot, "current_link_speed")
	data, err := os.ReadFile(linkSpeedFilePath)
	if err != nil {
		fmt.Printf("could not get link speed for slot %s\n", slot)
		return
	}
	value, err := getFloatFromLinkSpeed(string(data))
	if err != nil {
		fmt.Printf("Could not parse link speed from slot %s\n", slot)
		return
	}
	ch <- prometheus.MustNewConstMetric(collector.RegionMetric, prometheus.GaugeValue, value, slot)
	wg.Done()
}

func getFloatFromLinkSpeed(st string) (float64, error) {
	pattern := `([0-9.]+).*`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(st)

	// Check if a match is found
	if len(matches) < 2 {
		return 0, fmt.Errorf("no float value found in the input string\n")
	}

	// Extract and parse the float value
	floatValue, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse float value: %v\n", err)
	}

	return floatValue, nil
}
