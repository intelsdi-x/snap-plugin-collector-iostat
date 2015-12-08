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
	"io"
	"os"
	"strings"
	"time"

	"github.com/intelsdi-x/snap-plugin-collector-iostat/iostat/command"
	"github.com/intelsdi-x/snap-plugin-collector-iostat/iostat/parser"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/ctypes"
)

const (
	// Name of plugin
	Name = "iostat"
	// Version of plugin
	Version = 2
	// Type of plugin
	Type = plugin.CollectorPluginType
)

type runsCmd interface {
	Run(cmd string, args []string) (io.Reader, error)
}

type parses interface {
	Parse(io.Reader) ([]string, map[string]float64, error)
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
func (iostat *IOSTAT) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	_, data, err := iostat.run(mts)
	if err != nil {
		return nil, err
	}
	metrics := make([]plugin.PluginMetricType, len(mts))
	hostname, _ := os.Hostname()
	for i, m := range mts {
		if v, ok := data[joinNamespace(m.Namespace())]; ok {
			metrics[i] = plugin.PluginMetricType{
				Namespace_: m.Namespace(),
				Data_:      v,
				Source_:    hostname,
				Timestamp_: time.Now(),
			}
		}
	}
	return metrics, nil
}

// GetMetricTypes returns the metric types exposed by iostat
func (iostat *IOSTAT) GetMetricTypes(_ plugin.PluginConfigType) ([]plugin.PluginMetricType, error) {
	keys, _, err := iostat.run([]plugin.PluginMetricType{})
	if err != nil {
		return nil, err
	}
	mts := make([]plugin.PluginMetricType, len(keys))
	for i, k := range keys {
		mts[i] = plugin.PluginMetricType{Namespace_: strings.Split(strings.TrimPrefix(k, "/"), "/")}
	}
	return mts, nil
}

//GetConfigPolicy
func (iostat *IOSTAT) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	c := cpolicy.New()
	return c, nil
}

// Init initializes iostat plugin
func (iostat *IOSTAT) run(mts []plugin.PluginMetricType) ([]string, map[string]float64, error) {
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
func getArgs(mts []plugin.PluginMetricType) []string {
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

func joinNamespace(ns []string) string {
	return "/" + strings.Join(ns, "/")
}
