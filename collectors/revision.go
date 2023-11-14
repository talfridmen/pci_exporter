package collectors

import (
	"fmt"
	"os"
	"path/filepath"

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

func (collector *RevisionCollector) Collect(ch chan<- prometheus.Metric, slot string) {
	revisionFilePath := filepath.Join(PciDevicesPath, slot, "revision")
	if !fileExists(revisionFilePath) {
		return
	}
	data, err := os.ReadFile(revisionFilePath)
	if err != nil {
		fmt.Printf("could not get revisions for slot %s\n", slot)
		return
	}
	ch <- prometheus.MustNewConstMetric(collector.RevisionMetric, prometheus.GaugeValue, 1, slot, string(data[2:4]))
}
