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

	// for channeling
	"github.com/intelsdi-x/snap-plugin-utilities/source"
)

const (
	// Name of plugin
	Name = "iostat"
	// Version of plugin
	Version = 2
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
	execCommand(src source.Source, data chan interface{}, err chan error)
}

// IOSTAT
type IOSTAT struct {
	keys      []string
	data      map[string]interface{}
	timestamp time.Time
	mutex     *sync.RWMutex
	cmdCall   Command
}

type parser struct {
	// this structure is used in parsing iostat command output
	firstLine   bool // set true if next interval is exepected
	titleLine   bool // set true if the line is a title
	emptyTokens int  // numbers of empty tokens got from data channel

	statType    string   // type of statistics (for example cpu or device statistic)
	statSubType string   // subtype of statistics (for example sda)
	statNames   []string // names of statistics

	stat   string   // statictic namespace, includes statType, statSubType and StatName
	stats  []string // slice of statistics, after parsing process it's equivalent to IOSTAT.keys
	values []string // slice of statictics' values
}

type RealCmdCall struct{}

func (c *RealCmdCall) execCommand(src source.Source, data chan interface{}, err chan error) {
	// generate data (execute command defined in source struct) and put into buffered channels
	// see more in https://github.com/intelsdi-x/snap-plugin-utilities/tree/master/source
	src.Generate(data, err)
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

func (p *parser) process(data string, iostat *IOSTAT) error {
	line := removeEmptyStr(strings.Split(strings.TrimSpace(data), " "))
	if len(line) == 0 {
		// slice "line" is empty
		p.emptyTokens++
		if p.emptyTokens > defaultEmptyTokenAcceptance {
			return errors.New("more empty data slice than acceptance level")
		}
		return nil
	}
	p.emptyTokens = 0

	if p.titleLine {
		// skip the title line
		p.titleLine = false
		return nil
	}
	if p.firstLine {
		//next interval separated by a newline
		p.firstLine = false
		p.stats = []string{}
		p.values = []string{}
		/////////////////////////////////////////////////////////////////////////////
		// HERE: 	possible parsing output to get timestamp delivered by iostat
		//			the default format: "MM/DD/YYYY HH:MM:SS AM/PM"
		/////////////////////////////////////////////////////////////////////////////

		// getting timestamp as current local time
		// in format: "YYYY-MM-DD HH:MM:SS.fffffffff +/-HHMM TZ
		iostat.timestamp = time.Now()
		return nil
	}

	if strings.HasSuffix(line[0], ":") {
		if len(line) > 1 {
			p.statType = strings.ToLower(strings.TrimSuffix(line[0], ":"))
			p.statNames = replaceByPerSec(line[1:])
			return nil
		}
	}

	if len(p.statNames) == 0 || len(p.statType) == 0 {
		return errors.New("Invalid format of iostat output data")
	}
	if len(line) > len(p.statNames) {
		// subType is defined
		p.statSubType = line[0]
		line = line[1:]
	} else {
		p.statSubType = ""
	}

	if len(line) == len(p.statNames) && len(p.statNames) != 0 {
		for i, sname := range p.statNames {
			if p.statSubType != "" {
				p.stat = p.statType + "/" + p.statSubType + "/" + sname
			} else {
				p.stat = p.statType + "/" + sname
			}

			p.stats = append(p.stats, p.stat)
			p.values = append(p.values, line[i])
		}
	}

	if strings.ToLower(p.statSubType) == "all" {
		// all available metrics keys collected
		p.firstLine = true // for next scan skip first line

		if len(iostat.keys) == 0 {

			if len(p.stat) == 0 {
				return errors.New("can not retrive iostat metrics namespace")
			}

			iostat.mutex.Lock()
			iostat.keys = make([]string, len(p.stats))
			for i, s := range p.stats {
				iostat.keys[i] = joinNamespace(createNamespace(s))
			}
			iostat.mutex.Unlock()
		}

		if len(p.values) != len(iostat.keys) {
			// number of values has to be equivalent to number of keys
			return errors.New("invalid parsing iostat output")
		}

		iostat.mutex.Lock()
		for i, val := range p.values {
			v, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
			if err == nil {
				iostat.data[iostat.keys[i]] = v
			} else {
				fmt.Fprintln(os.Stderr, "invalid metric value", err)
				iostat.data[iostat.keys[i]] = nil
			}

		}
		iostat.mutex.Unlock()
	}

	return nil
}

// Init initializes iostat plugin
func (iostat *IOSTAT) init() (*IOSTAT, error) {

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
	var cmd string

	if path := iostat.cmdCall.getEnv("SNAP_IOSTAT_PATH"); path != "" {
		cmd = filepath.Join(path, "iostat")
	} else {
		path, err := iostat.cmdCall.lookPath("iostat")

		if err != nil {
			fmt.Fprint(os.Stderr, "Unable to find iostat (\"sysstat\" package).  Ensure it's in your path or set SNAP_IOSTAT_PATH.")
			panic(err)
		}
		cmd = path
	}

	// create source for command execution with args, see more in github.com/intelsdi-x/snap-plugin-utilities/source)
	src := source.Source{
		Command: cmd,
		Args:    args,
	}

	// create buffered channels for data and errors
	chan_data := make(chan interface{}, 100)
	chan_err := make(chan error, 100)

	// execute command with args defined in "source" structure, put data and err into buffered channels
	go iostat.cmdCall.execCommand(src, chan_data, chan_err)

	go func() {
		parser := &parser{firstLine: true, titleLine: true}

	LOOP:
		for {
			select {
			case data, ok := <-chan_data:
				if !ok {
					break LOOP
				}
				err := parser.process(data.(string), iostat)
				if err != nil {
					// print error in stderr file
					fmt.Fprintf(os.Stderr, "\nError: %+v\n", err)
					break LOOP
				}
			case e, ok := <-chan_err:
				if !ok {
					break LOOP
				}
				if e != nil {
					// print error in stderr file
					fmt.Fprintf(os.Stderr, "\nError: %+v\n", e)
					break LOOP
				}

			default:
				time.Sleep(1 * time.Microsecond)
			}
		}
	}()

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
