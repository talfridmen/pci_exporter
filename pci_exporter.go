package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"./collectors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define a struct for you collector that contains pointers
// to prometheus descriptors for each metric you wish to expose.
// Note you can also include fields of other types if they provide utility
// but we just won't be exposing them as metrics.
type PciCollector struct {
	PciDeviceMetric *prometheus.Desc
	driverNames     []string
	regionCollector regionCollector
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

// You must create a constructor for you collector that
// initializes every descriptor and returns a pointer to the collector
func newPciCollector() *PciCollector {
	driverFilter := flag.String("driver", "", "Specify the driver(s) to query (comma-separated)")
	flag.Parse()
	driverNames := strings.Split(*driverFilter, ",")

	return &PciCollector{
		PciDeviceMetric: prometheus.NewDesc("pci_device",
			"Describes information about PCI devices",
			[]string{"driver", "device", "slot", "revision", "link_speed", "link_width", "regions"}, nil,
		),
		driverNames: driverNames,
	}
}

// Each and every collector must implement the Describe function.
// It essentially writes all descriptors to the prometheus desc channel.
func (collector *PciCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.PciDeviceMetric
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// Collect implements required collect function for all promehteus collectors
func (collector *PciCollector) Collect(ch chan<- prometheus.Metric) {
	pciDriverPath := "/sys/bus/pci/drivers"
	drivers, err := os.ReadDir(pciDriverPath)
	if err != nil {
		fmt.Println("Error reading PCI drivers directory:", err)
		return
	}
	for _, driver := range drivers {
		// Check if the driver name is in the filter list (if specified)
		if len(collector.driverNames) > 0 && !contains(collector.driverNames, driver.Name()) {
			continue
		}

		driverInfo, err := collectDriverInfo(pciDriverPath, driver.Name())
		if err != nil {
			fmt.Println("Error collecting driver information:", err)
			continue
		}
		for _, deviceInfo := range driverInfo.devices {
			m1 := prometheus.MustNewConstMetric(collector.PciDeviceMetric, prometheus.GaugeValue, 1.0, driverInfo.name, deviceInfo.device, deviceInfo.slot, deviceInfo.revision, deviceInfo.link_speed, deviceInfo.link_width)
			//			m1 = prometheus.NewMetricWithTimestamp(time.Now(), m1)
			ch <- m1
		}
	}
}

func collectDriverInfo(pciDriverPath string, driverName string) (*DriverInfo, error) {
	driverPath := filepath.Join(pciDriverPath, driverName)

	slotsInfo, err := collectSlotsInfo(driverPath)
	if err != nil {
		return nil, err
	}

	return &DriverInfo{
		name:    driverName,
		devices: slotsInfo,
	}, nil
}

func collectSlotsInfo(driverPath string) ([]*DeviceInfo, error) {
	driverDir, err := os.ReadDir(driverPath)
	if err != nil {
		fmt.Printf("Error reading PCI driver directory for driver %s", driverPath)
		return nil, err
	}

	slots := []*DeviceInfo{}

	for _, slot := range driverDir {
		if strings.HasPrefix(slot.Name(), "0000") {
			deviceInfo, err := collectSlotInfo(driverPath, slot.Name())
			if err != nil {
				fmt.Printf("Error collecting slot info for slot %s", slot.Name())
				return nil, err
			}
			slots = append(slots, deviceInfo)
		}
	}

	return slots, nil
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
