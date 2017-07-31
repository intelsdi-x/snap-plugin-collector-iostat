/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package iostat

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/intelsdi-x/snap-plugin-collector-iostat/iostat/command"
	"github.com/intelsdi-x/snap-plugin-collector-iostat/iostat/parser"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	deviceMetric = "device"
)

type runsCmd interface {
	Run(cmd string, args []string) (io.Reader, error)
	Exec(cmd string, args []string) string
}

type parses interface {
	Parse(io.Reader) ([]string, map[string]float64, error)
	ParseVersion(string) ([]int64, error)
}

// Iostat structure
type Iostat struct {
	cmd    runsCmd
	parser parses
}

// NewIostatCollector returns instance of iostat object
func NewIostatCollector() *Iostat {
	return &Iostat{
		cmd:    command.New(),
		parser: parser.New(),
	}
}

// CollectMetrics returns metrics from iostat
func (iostat *Iostat) CollectMetrics(mts []plugin.Metric) ([]plugin.Metric, error) {
	_, data, err := iostat.run(mts)
	if err != nil {
		return nil, err
	}

	metrics := []plugin.Metric{}

	for _, mt := range mts {
		ns := mt.Namespace
		if len(ns) < 4 {
			return nil, fmt.Errorf("Namespace length is too short (len = %d)", len(ns))
		}

		if ns[3].Value == "*" {
			if ns[2].Value == deviceMetric {
				for k := range data {
					reg := ns[:2].String() + "/device/.*" + ns[4:].String()
					matched, err := regexp.MatchString(reg, k)
					if !matched {
						continue
					}
					if err != nil {
						return nil, fmt.Errorf("Error matching namespaces %v", ns)
					}

					dev, err := extractFromNamespace(k, 3)
					if err != nil {
						return nil, err
					}

					nsCopy := make(plugin.Namespace, len(ns))
					copy(nsCopy, ns)
					nsCopy[3].Value = dev

					if v, ok := data[nsCopy.String()]; ok {
						metrics = append(metrics, plugin.Metric{
							Namespace: nsCopy,
							Data:      v,
							Timestamp: time.Now(),
							Tags:      map[string]string{"dev": dev}})
					} else {
						fmt.Fprintf(os.Stdout, "No data found for metric %v", ns.Strings())
					}
				}
			} else {
				return nil, fmt.Errorf("Dynamic option * not supported for metric %v", ns)
			}
		} else {
			if v, ok := data[mt.Namespace.String()]; ok {
				metrics = append(metrics, plugin.Metric{
					Namespace: mt.Namespace,
					Data:      v,
					Timestamp: time.Now()})
			} else {
				fmt.Fprintf(os.Stdout, "No data found for metric %v", ns.Strings())
			}
		}
	}

	return metrics, nil
}

// GetMetricTypes returns the metric types exposed by iostat
func (iostat *Iostat) GetMetricTypes(_ plugin.Config) ([]plugin.Metric, error) {
	namespaces, _, err := iostat.run([]plugin.Metric{})
	if err != nil {
		return nil, err
	}

	mts := []plugin.Metric{}
	metric := plugin.Metric{}
	// List of terminal metric names
	mList := make(map[string]bool)
	for _, namespace := range namespaces {
		ns := plugin.NewNamespace(strings.Split(strings.TrimPrefix(namespace, "/"), "/")...)

		if len(ns) < 4 {
			return nil, fmt.Errorf("Namespace length is too short (len = %d)", len(ns))
		}
		// terminal metric name
		mItem := ns[len(ns)-1]
		if ns[2].Value == deviceMetric {
			if !mList[mItem.Value] {
				mList[mItem.Value] = true
				metric = plugin.Metric{
					Namespace: plugin.NewNamespace(parser.NsVendor, parser.NsType, deviceMetric).
						AddDynamicElement("device_id", "Device ID").
						AddStaticElement(mItem.Value),
					Description: "dynamic device metric: " + mItem.Value}
			} else {
				continue
			}
		} else {
			metric = plugin.Metric{Namespace: ns}
		}

		mts = append(mts, metric)
	}

	return mts, nil
}

//GetConfigPolicy return configuration policy
func (iostat *Iostat) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	c := plugin.NewConfigPolicy()
	return *c, nil

}

// Init initializes iostat plugin
func (iostat *Iostat) run(mts []plugin.Metric) ([]string, map[string]float64, error) {
	// TODO: allow the path and/or name of the command to be overriden through the pluginConfigType

	versionString := iostat.cmd.Exec("iostat", []string{"-V"})
	version, err := iostat.parser.ParseVersion(versionString)
	if err != nil {
		return nil, nil, err
	}
	if version[0] < 10 || (version[0] == 10 && version[1] < 2) {
		return nil, nil, fmt.Errorf("Iostat %d.%d.%d version (required >=10.2.0)", version[0], version[1], version[2])
	}

	reader, err := iostat.cmd.Run("iostat", getArgs(mts))
	if err != nil {
		return nil, nil, err
	}

	return iostat.parser.Parse(reader)
}

// updateArgs will add -y to the args provided to iostat telling iostat to report stats
// since the machine has booted if the config ReportSinceBoot is present and True.  The
// config for each metric being requested is the same so we need to check the config for
// one metric being requested.
func getArgs(mts []plugin.Metric) []string {
	/////////////////////////////////////////////////////////////////////////////////////////
	// 	IOstat command with interval 1 and options:
	// 		-c	 	display the CPU utilization report
	// 		-d	 	display the device utilization report
	// 		-p	 	display statistics for block devices and all their partitions
	// 		-g ALL	display statistics for a group of devices
	// 		-x		display extended statistics
	// 		-k		display bandwidth statistics in kilobytes per second
	// 		-t		print the time for each report displayed
	//      -y      excludes report since last report (not since boot)
	////////////////////////////////////////////////////////////////////////////////////////
	iostatArgs := []string{"-c", "-d", "-p", "-g", "ALL", "-x", "-k", "-t"}

	reportLatest := true
	if len(mts) > 0 && mts[0].Config != nil && len(mts[0].Config) > 0 {
		if m, ok := mts[0].Config["ReportSinceBoot"]; ok {
			switch val := m.(type) {
			case plugin.Config:
				if value, err := val.GetBool("ReportSinceBoot"); err != nil {
					// return without adding '-y' to the args
					// produces results since boot
					if value {
						reportLatest = false
					}
				}
			}
		}
	}
	if reportLatest {
		// -y will disregards the summary since boot
		// the "1" "1" arguments will produce 1 result over a 1 second interval
		iostatArgs = append(iostatArgs, "-y", "1", "1")
	}
	return iostatArgs
}

// extractFromNamespace extracts element of index i from namespace string
func extractFromNamespace(namespace string, i int) (string, error) {
	ns := plugin.NewNamespace(strings.Split(strings.TrimPrefix(namespace, "/"), "/")...)
	if len(ns) < i+1 {
		return "", fmt.Errorf("Cannot extract element from namespace, index out of range (i = %d)", i)
	}
	return ns[i].Value, nil
}
