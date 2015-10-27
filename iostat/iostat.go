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
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
)

const (
	// Name of plugin
	Name = "iostat"
	// Version of plugin
	Version = 1
	// Type of plugin
	Type = plugin.CollectorPluginType

	ns_vendor = "intel"
	ns_class  = "linux"
	ns_type   = "iostat"

	defaultTimeout              = 2 * time.Second
	defaultEmptyTokenAcceptance = 5
)

//interfaces needed for mocking command output in iostat_test.go
type Command interface {
	getEnv(string) string
	lookPath(string) (string, error)
	execCommand(name string, arg ...string) *exec.Cmd
}

// IOSTAT
type IOSTAT struct {
	keys      []string
	data      map[string]interface{}
	timestamp time.Time
	mutex     *sync.RWMutex
	cmdCall   Command
}

type RealCmdCall struct{}

// execCommand returns the Cmd struct to execute the named program with the given arguments
func (c *RealCmdCall) execCommand(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

func (c *RealCmdCall) getEnv(key string) string {
	return os.Getenv(key)
}

func (c *RealCmdCall) lookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// CollectMetrics returns metrics from iostat
func (iostat *IOSTAT) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	metrics := make([]plugin.PluginMetricType, len(mts))
	iostat.mutex.RLock()
	defer iostat.mutex.RUnlock()
	hostname, _ := os.Hostname()
	for i, m := range mts {
		if v, ok := iostat.data[joinNamespace(m.Namespace())]; ok {
			metrics[i] = plugin.PluginMetricType{
				Namespace_: m.Namespace(),
				Data_:      v,
				Source_:    hostname,
				Timestamp_: iostat.timestamp,
			}
		}
	}
	return metrics, nil
}

// GetMetricTypes returns the metric types exposed by iostat
func (iostat *IOSTAT) GetMetricTypes(_ plugin.PluginConfigType) ([]plugin.PluginMetricType, error) {
	mts := make([]plugin.PluginMetricType, len(iostat.keys))
	iostat.mutex.RLock()
	defer iostat.mutex.RUnlock()
	for i, k := range iostat.keys {
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
func (iostat *IOSTAT) init() (*IOSTAT, error) {
	var cmd *exec.Cmd

	/////////////////////////////////////////////////////////////////////////////////////////
	// 	IOstat command with interval 1 and options:
	// 		-c	 	display the CPU utilization report
	// 		-d	 	display the device utilization report
	// 		-p	 	display statistics for block devices and all their partitions
	// 		-g ALL	display statistics for a group of devices
	// 		-x		display extended statistics
	// 		-k		display bandwidth statistics in kilobytes per second
	// 		-t		print the time for each report displayed
	////////////////////////////////////////////////////////////////////////////////////////

	args := []string{"-c", "-d", "-p", "-g", "ALL", "-x", "-k", "-t", "1"}

	if path := iostat.cmdCall.getEnv("SNAP_IOSTAT_PATH"); path != "" {
		cmd = iostat.cmdCall.execCommand(filepath.Join(path, "iostat"), args...)
	} else {
		c, err := iostat.cmdCall.lookPath("iostat")
		if err != nil {
			fmt.Fprint(os.Stderr, "Unable to find iostat (\"systat\" package).  Ensure it's in your path or set SNAP_IOSTAT_PATH.")
			panic(err)
		}
		cmd = iostat.cmdCall.execCommand(c, args...)
	}
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe", err)
		return nil, err
	}

	// read the data from stdout
	scanner := bufio.NewScanner(cmdReader)
	title := true
	emptyTokens := 0

	go func() {
		first := true
		var statType string
		var statSubType string
		var statNames []string
		var keys []string
		var data []string
		var key string

		for scanner.Scan() {
			line := removeEmptyStr(strings.Split(strings.TrimSpace(scanner.Text()), " "))
			if len(line) == 0 {
				// slice "line" is empty
				emptyTokens++
				if emptyTokens > defaultEmptyTokenAcceptance {
					// more empty slice than acceptance level means error occurs, break the loop
					break
				}
				continue
			}
			emptyTokens = 0

			if title {
				// skip the title line
				title = false
				continue
			}
			if first {
				//next interval separated by a newline
				first = false
				keys = []string{}
				data = []string{}

				/////////////////////////////////////////////////////////////////////////////
				// HERE: 	possible parsing output to get timestamp delivered by iostat
				//			the default format: "MM/DD/YYYY HH:MM:SS AM/PM"
				/////////////////////////////////////////////////////////////////////////////

				// getting timestamp as current local time
				// in format: "YYYY-MM-DD HH:MM:SS.fffffffff +/-HHMM TZ
				iostat.timestamp = time.Now()
				continue
			}

			if strings.HasSuffix(line[0], ":") {
				if len(line) > 1 {
					statType = strings.ToLower(strings.TrimSuffix(line[0], ":"))
					statNames = replaceByPerSec(line[1:])
					continue
				}
			}

			if len(statNames) == 0 || len(statType) == 0 {
				fmt.Fprintf(os.Stderr, "Invalid format of iostat output data\n")
				break
			}
			if len(line) > len(statNames) {
				// subType is defined
				statSubType = line[0]
				line = line[1:]
			} else {
				statSubType = ""
			}

			if len(line) == len(statNames) && len(statNames) != 0 {
				for i, statName := range statNames {

					if statSubType != "" {
						key = statType + "/" + statSubType + "/" + statName
					} else {
						key = statType + "/" + statName
					}

					keys = append(keys, key)
					data = append(data, line[i])
				}
			}

			if strings.ToLower(statSubType) == "all" {
				// all available metrics keys collected
				first = true // for next scan skip first line

				if len(iostat.keys) == 0 {

					if len(keys) == 0 {
						panic(errors.New("Error getting iostat metrics namespace"))
					}

					iostat.mutex.Lock()
					iostat.keys = make([]string, len(keys))
					for i, k := range keys {
						iostat.keys[i] = joinNamespace(createNamespace(k))
					}
					iostat.mutex.Unlock()
				}

				if len(data) != len(iostat.keys) {
					panic(errors.New("Invalid parsing iostat output"))
				}

				iostat.mutex.Lock()
				for i, d := range data {
					v, err := strconv.ParseFloat(strings.TrimSpace(d), 64)
					if err == nil {
						iostat.data[iostat.keys[i]] = v
					} else {
						fmt.Fprintln(os.Stderr, "Invalid metric value", err)
						iostat.data[iostat.keys[i]] = nil
					}

				}
				iostat.mutex.Unlock()
			}
		} // end of scanning
	}() // end of go routine

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting iostat", err)
		return nil, fmt.Errorf(fmt.Sprintf("Error starting iostat command, err=%+v", err))
	}
	// we need to wait until we have our metric types
	st := time.Now()
	for {
		iostat.mutex.RLock()
		c := len(iostat.keys)
		iostat.mutex.RUnlock()
		if c > 0 {
			break
		}
		if time.Since(st) > defaultTimeout {
			return nil, fmt.Errorf("Timed out waiting for metrics from iostat")
		}
	}
	return iostat, nil
}

func joinNamespace(ns []string) string {
	return "/" + strings.Join(ns, "/")
}

// createNamespace returns namespace slice of strings composed from: vendor, class, type and ceph-daemon name
func createNamespace(name string) []string {
	return []string{ns_vendor, ns_class, ns_type, name}
}

// removeEmptyStr removes empty strings from slice
func removeEmptyStr(slice []string) []string {
	var result []string
	for _, str := range slice {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
}

// replacePerSec turns "/s" into "_per_sec"
func replaceByPerSec(slice []string) []string {
	for i, str := range slice {
		slice[i] = strings.Replace(str, "/s", "_per_sec", 1)
	}
	return slice
}

// New returns snap-plugin-collector-iostat instance
func New() (*IOSTAT, error) {
	iostat := &IOSTAT{mutex: &sync.RWMutex{}, data: map[string]interface{}{}, cmdCall: &RealCmdCall{}}
	return iostat.init()
}
