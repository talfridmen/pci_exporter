package collectors

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type RevisionCollector struct {
	RegionMetric *prometheus.Desc
}

func NewRevisionCollector() *RevisionCollector {
	return &RevisionCollector{
		RegionMetric: prometheus.NewDesc(
			"pci_device_revision",
			"The revisions of the pci device",
			[]string{"device", "revision"},
			nil,
		),
	}
}

func (collector *RevisionCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.RegionMetric
}

func (collector *RevisionCollector) Collect(wg *sync.WaitGroup, ch chan<- prometheus.Metric, slot string) {
	revisionFilePath := filepath.Join(PciDevicesPath, slot, "revision")
	data, err := os.ReadFile(revisionFilePath)
	if err != nil {
		fmt.Printf("could not get revisions for slot %s", slot)
		return
	}
	ch <- prometheus.MustNewConstMetric(collector.RegionMetric, prometheus.GaugeValue, 1, slot, string(data))
	wg.Done()
}
