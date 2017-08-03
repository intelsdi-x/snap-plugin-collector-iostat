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
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
)

const (
	// Name of plugin
	Name = "iostat"
	// Version of plugin
	Version = 6
	// Type of plugin
	Type         = plugin.CollectorPluginType
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

// IOSTAT
type IOSTAT struct {
	cmd    runsCmd
	parser parses
}

// New returns snap-plugin-collector-iostat instance
func New() (*IOSTAT, error) {
	iostat := &IOSTAT{
		cmd:    command.New(),
		parser: parser.New(),
	}
	return iostat, nil
}

// CollectMetrics returns metrics from iostat
func (iostat *IOSTAT) CollectMetrics(mts []plugin.MetricType) ([]plugin.MetricType, error) {
	_, data, err := iostat.run(mts)
	if err != nil {
		return nil, err
	}

	metrics := []plugin.MetricType{}

	for _, mt := range mts {
		ns := mt.Namespace()
		if len(ns) < 4 {
			return nil, fmt.Errorf("Namespace length is too short (len = %d)", len(ns))
		}

		if ns[3].Value == "*" {
			if ns[2].Value == deviceMetric {
				for k, _ := range data {
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

					nsCopy := make(core.Namespace, len(ns))
					copy(nsCopy, ns)
					nsCopy[3].Value = dev

					if v, ok := data[nsCopy.String()]; ok {
						metrics = append(metrics, plugin.MetricType{
							Namespace_: nsCopy,
							Data_:      v,
							Timestamp_: time.Now(),
							Tags_:      map[string]string{"dev": dev}})
					} else {
						fmt.Fprintf(os.Stdout, "No data found for metric %v", ns.Strings())
					}
				}
			} else {
				return nil, fmt.Errorf("Dynamic option * not supported for metric %v", ns)
			}
		} else {
			if v, ok := data[mt.Namespace().String()]; ok {
				metrics = append(metrics, plugin.MetricType{
					Namespace_: mt.Namespace(),
					Data_:      v,
					Timestamp_: time.Now()})
			} else {
				fmt.Fprintf(os.Stdout, "No data found for metric %v", ns.Strings())
			}
		}
	}

	return metrics, nil
}

// GetMetricTypes returns the metric types exposed by iostat
func (iostat *IOSTAT) GetMetricTypes(_ plugin.ConfigType) ([]plugin.MetricType, error) {
	namespaces, _, err := iostat.run([]plugin.MetricType{})
	if err != nil {
		return nil, err
	}

	mts := []plugin.MetricType{}
	metric := plugin.MetricType{}
	// List of terminal metric names
	mList := make(map[string]bool)
	for _, namespace := range namespaces {
		ns := core.NewNamespace(strings.Split(strings.TrimPrefix(namespace, "/"), "/")...)

		if len(ns) < 4 {
			return nil, fmt.Errorf("Namespace length is too short (len = %d)", len(ns))
		}
		// terminal metric name
		mItem := ns[len(ns)-1]
		if ns[2].Value == deviceMetric {
			if !mList[mItem.Value] {
				mList[mItem.Value] = true
				metric = plugin.MetricType{
					Namespace_: core.NewNamespace(parser.NsVendor, parser.NsType, deviceMetric).
						AddDynamicElement("device_id", "Device ID").
						AddStaticElement(mItem.Value),
					Description_: "dynamic device metric: " + mItem.Value}
			} else {
				continue
			}
		} else {
			metric = plugin.MetricType{Namespace_: ns}
		}

		mts = append(mts, metric)
	}

	return mts, nil
}

//GetConfigPolicy
func (iostat *IOSTAT) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	c := cpolicy.New()
	return c, nil
}

// Init initializes iostat plugin
func (iostat *IOSTAT) run(mts []plugin.MetricType) ([]string, map[string]float64, error) {
	// TODO: reminder - remove these todo statements from the README (roadmap section) once
	//       they are completed
	// TODO: add validation that we are running sysstat version 10.2.0 or greater else print a
	//       a warning or otherwise communicate to the user that they are potentially using an
	//       unsupported/tested configuration
	// TODO: allow the path and/or name of the command to be overriden through the pluginConfigType

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
func getArgs(mts []plugin.MetricType) []string {
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
	if len(mts) > 0 && mts[0].Config() != nil && len(mts[0].Config().Table()) > 0 {
		if m, ok := mts[0].Config().Table()["ReportSinceBoot"]; ok {
			switch val := m.(type) {
			case ctypes.ConfigValueBool:
				if val.Value {
					// return without adding '-y' to the args
					// produces results since boot
					reportLatest = false
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
	ns := core.NewNamespace(strings.Split(strings.TrimPrefix(namespace, "/"), "/")...)
	if len(ns) < i+1 {
		return "", fmt.Errorf("Cannot extract element from namespace, index out of range (i = %d)", i)
	}
	return ns[i].Value, nil
}
