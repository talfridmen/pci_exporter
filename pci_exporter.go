package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	PciDeviceMetric *prometheus.Desc
	driverNames     []string
	regionCollector *collectors.RegionCollector
}

type DeviceInfo struct {
	device     string
	slot       string
	revision   string
	link_speed string
	link_width string
}

type DriverInfo struct {
	devices []*DeviceInfo
	name    string
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

// You must create a constructor for you collector that
// initializes every descriptor and returns a pointer to the collector
func newPciCollector() *PciCollector {
	driverFilter := flag.String("driver", "", "Specify the driver(s) to query (comma-separated)")
	flag.Parse()
	driverNames := strings.Split(*driverFilter, ",")
	driverNames = filter(
		driverNames,
		func(st string) bool {
			return st != ""
		},
	)

	return &PciCollector{
		PciDeviceMetric: prometheus.NewDesc("pci_device",
			"Describes information about PCI devices",
			[]string{"driver", "device", "slot", "revision", "link_speed", "link_width", "regions"}, nil,
		),
		driverNames:     driverNames,
		regionCollector: collectors.NewRegionCollector(),
	}
}

// Each and every collector must implement the Describe function.
// It essentially writes all descriptors to the prometheus desc channel.
func (collector *PciCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.PciDeviceMetric
	collector.regionCollector.Describe(ch)
}

// Collect implements required collect function for all promehteus collectors
func (collector *PciCollector) Collect(ch chan<- prometheus.Metric) {
	slots := []string{}
	if len(collector.driverNames) > 0 {
		drivers, err := os.ReadDir(PciDriversPath)
		if err != nil {
			fmt.Println("Error reading PCI drivers directory:", err)
			return
		}
		for _, driver := range drivers {
			// Check if the driver name is in the filter list (if specified)
			if len(collector.driverNames) > 0 && !contains(collector.driverNames, driver.Name()) {
				continue
			}

			driverPath := filepath.Join(PciDriversPath, driver.Name())

			driverDirElements, err := os.ReadDir(driverPath)
			if err != nil {
				fmt.Printf("Could not ls driver directory for driver %s", driver.Name())
				return
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
			fmt.Println("Error reading PCI drivers directory:", err)
			return
		}
		for _, device := range devices {
			slots = append(slots, device.Name())
		}
	}

	for _, slot := range slots {
		collector.regionCollector.Collect(ch, slot)
	}
}

func collectSlotInfo(driverPath string, slot string) (*DeviceInfo, error) {
	slotPath := filepath.Join(driverPath, slot)

	revision, err := os.ReadFile(filepath.Join(slotPath, "revision"))
	if err != nil {
		fmt.Printf("Error reading revision for slot %s", slot)
		return nil, err
	}
	link_speed, err := os.ReadFile(filepath.Join(slotPath, "current_link_speed"))
	if err != nil {
		fmt.Printf("Error reading link_speed for slot %s", slot)
		return nil, err
	}
	link_width, err := os.ReadFile(filepath.Join(slotPath, "current_link_width"))
	if err != nil {
		fmt.Printf("Error reading link_width for slot %s", slot)
		return nil, err
	}
	device, err := os.ReadFile(filepath.Join(slotPath, "device"))
	if err != nil {
		fmt.Printf("Error reading device for slot %s", slot)
		return nil, err
	}
	vendor, err := os.ReadFile(filepath.Join(slotPath, "vendor"))
	if err != nil {
		fmt.Printf("Error reading vendor for slot %s", slot)
		return nil, err
	}

	return &DeviceInfo{
		device:     strings.Join([]string{strings.TrimSpace(string(vendor[2:])), strings.TrimSpace(string(device[2:]))}, ":"),
		slot:       slot,
		revision:   strings.TrimSpace(string(revision))[2:],
		link_speed: strings.TrimSpace(string(link_speed)),
		link_width: strings.TrimSpace(string(link_width)),
	}, nil
}

func main() {
	collector := newPciCollector()
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe("0.0.0.0:9101", nil))
}
