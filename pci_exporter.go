package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/talfridmen/pci_exporter/collectors"
)

const (
	PciDevicesPath = "/sys/bus/pci/devices/"
	PciDriversPath = "/sys/bus/pci/drivers/"
)

// Define a struct for you collector that contains pointers
// to prometheus descriptors for each metric you wish to expose.
// Note you can also include fields of other types if they provide utility
// but we just won't be exposing them as metrics.
type PciCollector struct {
	driverNames        []string
	regionCollector    *collectors.RegionCollector
	revisionCollector  *collectors.RevisionCollector
	linkSpeedCollector *collectors.LinkSpeedCollector
	linkWidthCollector *collectors.LinkWidthCollector
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func filter(arr []string, cond func(string) bool) []string {
	result := []string{}
	for i := range arr {
		if cond(arr[i]) {
			result = append(result, arr[i])
		}
	}
	return result
}

func getSlots(driverNames []string) []string {
	slots := []string{}
	if len(driverNames) > 0 {
		drivers, err := os.ReadDir(PciDriversPath)
		if err != nil {
			fmt.Printf("Error reading PCI drivers directory: %v\n", err)
			return nil
		}
		for _, driver := range drivers {
			// Check if the driver name is in the filter list (if specified)
			if len(driverNames) > 0 && !contains(driverNames, driver.Name()) {
				continue
			}

			driverPath := filepath.Join(PciDriversPath, driver.Name())

			driverDirElements, err := os.ReadDir(driverPath)
			if err != nil {
				fmt.Printf("Could not ls driver directory for driver %s\n", driver.Name())
				return nil
			}

			for _, element := range driverDirElements {
				if strings.HasPrefix(element.Name(), "0000") {
					slots = append(slots, element.Name())
				}
			}
		}
	} else {
		devices, err := os.ReadDir(PciDevicesPath)
		if err != nil {
			fmt.Printf("Error reading PCI drivers directory: %v\n", err)
			return nil
		}
		for _, device := range devices {
			slots = append(slots, device.Name())
		}
	}
	return slots
}

func getDriverNames() []string {
	driverFilter := flag.String("driver", "", "Specify the driver(s) to query (comma-separated)")
	flag.Parse()
	driverNames := strings.Split(*driverFilter, ",")
	driverNames = filter(
		driverNames,
		func(st string) bool {
			return st != ""
		},
	)
	return driverNames
}

// You must create a constructor for you collector that
// initializes every descriptor and returns a pointer to the collector
func newPciCollector() *PciCollector {
	return &PciCollector{
		driverNames:        getDriverNames(),
		regionCollector:    collectors.NewRegionCollector(),
		revisionCollector:  collectors.NewRevisionCollector(),
		linkSpeedCollector: collectors.NewLinkSpeedCollector(),
		linkWidthCollector: collectors.NewLinkWidthCollector(),
	}
}

// Each and every collector must implement the Describe function.
// It essentially writes all descriptors to the prometheus desc channel.
func (collector *PciCollector) Describe(ch chan<- *prometheus.Desc) {
	collector.regionCollector.Describe(ch)
	collector.revisionCollector.Describe(ch)
	collector.linkSpeedCollector.Describe(ch)
	collector.linkWidthCollector.Describe(ch)
}

// Collect implements required collect function for all promehteus collectors
func (collector *PciCollector) Collect(ch chan<- prometheus.Metric) {
	slots := getSlots(collector.driverNames)

	var wg sync.WaitGroup
	for _, slot := range slots {
		wg.Add(4)
		go collector.regionCollector.Collect(&wg, ch, slot)
		go collector.revisionCollector.Collect(&wg, ch, slot)
		go collector.linkSpeedCollector.Collect(&wg, ch, slot)
		go collector.linkWidthCollector.Collect(&wg, ch, slot)
	}
	wg.Wait()
}

// device, err := os.ReadFile(filepath.Join(slotPath, "device"))
// if err != nil {
// 	fmt.Printf("Error reading device for slot %s", slot)
// 	return nil, err
// }
// vendor, err := os.ReadFile(filepath.Join(slotPath, "vendor"))
// if err != nil {
// 	fmt.Printf("Error reading vendor for slot %s", slot)
// 	return nil, err
// }

func main() {
	collector := newPciCollector()
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe("0.0.0.0:9101", nil))
}
