package collectors

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type RevisionCollector struct {
	RevisionMetric *prometheus.Desc
}

func NewRevisionCollector() *RevisionCollector {
	return &RevisionCollector{
		RevisionMetric: prometheus.NewDesc(
			"pci_device_revision",
			"The revisions of the pci device",
			[]string{"device", "revision"},
			nil,
		),
	}
}

func (collector *RevisionCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.RevisionMetric
}

func (collector *RevisionCollector) Collect(wg *sync.WaitGroup, ch chan<- prometheus.Metric, slot string) {
	revisionFilePath := filepath.Join(PciDevicesPath, slot, "revision")
	if !fileExists(revisionFilePath) {
		wg.Done()
		return
	}
	data, err := os.ReadFile(revisionFilePath)
	if err != nil {
		fmt.Printf("could not get revisions for slot %s\n", slot)
		wg.Done()
		return
	}
	ch <- prometheus.MustNewConstMetric(collector.RevisionMetric, prometheus.GaugeValue, 1, slot, string(data))
	wg.Done()
}
