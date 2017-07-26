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
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/snap-plugin-collector-iostat/iostat/parser"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
)

// type Mock struct {
// 	key   string
// 	value float64
// }

// var ns_prefix = []string{parser.NsVendor,  parser.NsType}

// var mockKV = []Mock{
// 	{"/intel/iostat/avg-cpu/%user", 0.50},
// 	{"/intel/iostat/avg-cpu/%nice", 0.00},
// 	{"/intel/iostat/avg-cpu/%system", 0.13},
// 	{"/intel/iostat/avg-cpu/%iowait", 0.00},
// 	{"/intel/iostat/avg-cpu/%steal", 0.00},
// 	{"/intel/iostat/avg-cpu/%idle", 99.37},
// 	{"/intel/iostat/device/sda/rrqm_per_sec", 0.00},
// 	{"/intel/iostat/device/sdb/wrqm_per_sec", 0.33},
// 	{"/intel/iostat/device/sda1/r_per_sec", 0.00},
// 	{"/intel/iostat/device/sdb1/w_per_sec", 0.08},
// 	{"/intel/iostat/device/sda2/rkB_per_sec", 0.00},
// 	{"/intel/iostat/device/sdb2/wkB_per_sec", 4.55},
// 	{"/intel/iostat/device/sda3/avgrq-sz", 8.00},
// 	{"/intel/iostat/device/sda4/avgqu-sz", 0.00},
// 	{"/intel/iostat/device/sdb/await", 1.83},
// 	{"/intel/iostat/device/sdb/r_await", 0.94},
// 	{"/intel/iostat/device/sdb/w_await", 2.00},
// 	{"/intel/iostat/device/sdb/svctm", 0.06},
// 	{"/intel/iostat/device/sdb/%util", 0.00},
// 	{"/intel/iostat/device/ALL/rrqm_per_sec", 0.05},
// 	{"/intel/iostat/device/ALL/wkB_per_sec", 30.68},
// }

var refMap = map[string]interface{}{
	"/intel/iostat/device/sda1/%util":        0,
	"/intel/iostat/device/sda4/await":        0.11,
	"/intel/iostat/device/sdb1/r_per_sec":    0.04,
	"/intel/iostat/device/sda/wrqm_per_sec":  0,
	"/intel/iostat/device/ALL/avgrq-sz":      45.65,
	"/intel/iostat/device/sda4/avgrq-sz":     8,
	"/intel/iostat/device/sdb2/avgqu-sz":     0,
	"/intel/iostat/device/sda4/avgqu-sz":     0,
	"/intel/iostat/device/ALL/%util":         0,
	"/intel/iostat/device/sda/r_per_sec":     0,
	"/intel/iostat/device/sdb/avgrq-sz":      45.7,
	"/intel/iostat/device/sda/avgqu-sz":      0,
	"/intel/iostat/device/ALL/await":         1.82,
	"/intel/iostat/device/sda3/w_per_sec":    0,
	"/intel/iostat/device/ALL/avgqu-sz":      0,
	"/intel/iostat/device/sda3/rrqm_per_sec": 0,
	"/intel/iostat/device/sdb1/wrqm_per_sec": 0.07,
	"/intel/iostat/device/sda4/r_per_sec":    0,
	"/intel/iostat/device/sdb1/rrqm_per_sec": 0,
	"/intel/iostat/device/sda4/w_per_sec":    0,
	"/intel/iostat/device/sdb/w_per_sec":     0.64,
	"/intel/iostat/device/sdb/%util":         0,
	"/intel/iostat/device/sdb2/%util":        0,
	"/intel/iostat/device/sdb1/await":        9.81,
	"/intel/iostat/device/sda1/rrqm_per_sec": 0,
	"/intel/iostat/device/sdb/wrqm_per_sec":  0.33,
	"/intel/iostat/device/sdb2/r_per_sec":    0.09,
	"/intel/iostat/device/sda1/avgqu-sz":     0,
	"/intel/iostat/device/sdb1/%util":        0,
	"/intel/iostat/device/sda4/rrqm_per_sec": 0,
	"/intel/iostat/device/sda/w_per_sec":     0,
	"/intel/iostat/device/sda2/w_per_sec":    0,
	"/intel/iostat/device/sda1/w_per_sec":    0,
	"/intel/iostat/device/sdb1/avgrq-sz":     185.22,
	"/intel/iostat/device/sda1/await":        0.12,
	"/intel/iostat/device/ALL/rrqm_per_sec":  0.05,
	"/intel/iostat/device/sda4/wrqm_per_sec": 0,
	"/intel/iostat/device/sdb2/w_per_sec":    0.55,
	"/intel/iostat/device/sda2/avgrq-sz":     7.8,
	"/intel/iostat/device/sdb2/rrqm_per_sec": 0.02,
	"/intel/iostat/device/sdb/r_per_sec":     0.13,
	"/intel/iostat/device/sdb2/avgrq-sz":     19.87,
	"/intel/iostat/device/sda3/avgqu-sz":     0,
	"/intel/iostat/device/sda4/%util":        0,
	"/intel/iostat/device/sda/await":         0.1,
	"/intel/iostat/device/sda2/wrqm_per_sec": 0,
	"/intel/iostat/device/sda3/avgrq-sz":     8,
	"/intel/iostat/device/sdb1/avgqu-sz":     0,
	"/intel/iostat/device/sdb/rrqm_per_sec":  0.02,
	"/intel/iostat/device/sdb2/wrqm_per_sec": 0.26,
	"/intel/iostat/device/sdb1/w_per_sec":    0.08,
	"/intel/iostat/device/sda2/r_per_sec":    0,
	"/intel/iostat/device/ALL/w_per_sec":     1.27,
	"/intel/iostat/device/sdb/await":         1.83,
	"/intel/iostat/device/sda2/await":        0.08,
	"/intel/iostat/device/sda2/rrqm_per_sec": 0,
	"/intel/iostat/device/ALL/wrqm_per_sec":  0.66,
	"/intel/iostat/device/sda3/r_per_sec":    0,
	"/intel/iostat/device/ALL/r_per_sec":     0.26,
	"/intel/iostat/device/sdb/avgqu-sz":      0,
	"/intel/iostat/device/sda/%util":         0,
	"/intel/iostat/device/sdb2/await":        0.34,
	"/intel/iostat/device/sda3/wrqm_per_sec": 0,
	"/intel/iostat/device/sda/avgrq-sz":      8.06,
	"/intel/iostat/device/sda1/avgrq-sz":     8.19,
	"/intel/iostat/device/sda2/avgqu-sz":     0,
	"/intel/iostat/device/sda3/%util":        0,
	"/intel/iostat/device/sda2/%util":        0,
	"/intel/iostat/device/sda3/await":        0.12,
	"/intel/iostat/device/sda/rrqm_per_sec":  0,
	"/intel/iostat/device/sda1/wrqm_per_sec": 0,
	"/intel/iostat/device/sda1/r_per_sec":    0,
}
var mockExecOut = `sysstat version 11.2.0
(C) Sebastien Godard (sysstat <at> orange.fr)
`
var mockCmdOut = `Linux 3.10.0-229.11.1.el7.x86_64 (gklab-108-166) 0/26/2015      _x86_64_        (8 CPU)

			10/26/2015 06:36:57 AM
			avg-cpu:  %user   %nice %system %iowait  %steal   %idle
			           0.50    0.00    0.13    0.00    0.00   99.37

			Device:         rrqm/s   wrqm/s     r/s     w/s    rkB/s    wkB/s avgrq-sz avgqu-sz   await r_await w_await  svctm  %util
			sda               0.00     0.00    0.00    0.00     0.01     0.00     8.06     0.00    0.10    0.10    0.00   0.04   0.00
			sda1              0.00     0.00    0.00    0.00     0.00     0.00     8.19     0.00    0.12    0.12    0.00   0.12   0.00
			sda2              0.00     0.00    0.00    0.00     0.00     0.00     7.80     0.00    0.08    0.08    0.00   0.08   0.00
			sda3              0.00     0.00    0.00    0.00     0.00     0.00     8.00     0.00    0.12    0.12    0.00   0.12   0.00
			sda4              0.00     0.00    0.00    0.00     0.00     0.00     8.00     0.00    0.11    0.11    0.00   0.11   0.00
			sdb               0.02     0.33    0.13    0.64     2.08    15.34    45.70     0.00    1.83    0.94    2.00   0.06   0.00
			sdb1              0.00     0.07    0.04    0.08     0.26    10.79   185.22     0.00    9.81    0.23   14.21   0.25   0.00
			sdb2              0.02     0.26    0.09    0.55     1.81     4.55    19.87     0.00    0.34    1.24    0.20   0.03   0.00
			 ALL              0,05     0,66    0,26    1,27     4,17    30,68    45,65     0,00    1,82    0,92    2,00   0,06   0,00

		`

var staticMockMts = []plugin.MetricType{
	plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "iostat", "avg-cpu", "%user"),
	},
	plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "iostat", "avg-cpu", "%nice"),
	},
	plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "iostat", "avg-cpu", "%system"),
	},
	plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "iostat", "avg-cpu", "%iowait"),
	},
	plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "iostat", "avg-cpu", "%steal"),
	},
	plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "iostat", "avg-cpu", "%idle"),
	},
}

var dynamicMockMts = []plugin.MetricType{
	plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "iostat", "device").
			AddDynamicElement("device_id", "Device ID").
			AddStaticElement("%util"),
	},
	plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "iostat", "device").
			AddDynamicElement("device_id", "Device ID").
			AddStaticElement("await"),
	},
	plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "iostat", "device").
			AddDynamicElement("device_id", "Device ID").
			AddStaticElement("rrqm_per_sec"),
	},
	plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "iostat", "device").
			AddDynamicElement("device_id", "Device ID").
			AddStaticElement("wrqm_per_sec"),
	},
	plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "iostat", "device").
			AddDynamicElement("device_id", "Device ID").
			AddStaticElement("r_per_sec"),
	},
	plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "iostat", "device").
			AddDynamicElement("device_id", "Device ID").
			AddStaticElement("w_per_sec"),
	},
	plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "iostat", "device").
			AddDynamicElement("device_id", "Device ID").
			AddStaticElement("avgrq-sz"),
	},
	plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "iostat", "device").
			AddDynamicElement("device_id", "Device ID").
			AddStaticElement("avgqu-sz"),
	},
}

type mockCmdRunner struct{}

func (c *mockCmdRunner) Run(cmd string, args []string) (io.Reader, error) {
	return strings.NewReader(mockCmdOut), nil
}
func (c *mockCmdRunner) Exec(cmd string, args []string) string {
	return mockExecOut
}

//////////////////////////////////////////////////////////////////////////////
//	***						TESTS										***	//
//////////////////////////////////////////////////////////////////////////////

func TestIostat(t *testing.T) {
	iostat := &IOSTAT{parser: parser.New(), cmd: &mockCmdRunner{}}

	Convey("Given invalid metric namespace collect metrics", t, func() {
		badMetrics := []plugin.MetricType{
			plugin.MetricType{
				Namespace_: core.NewNamespace("intel", "iostat", "device", "sda", "bad"),
			},
		}
		So(func() { iostat.CollectMetrics(badMetrics) }, ShouldNotPanic)
		result, err := iostat.CollectMetrics(badMetrics)
		So(err, ShouldBeNil)

		for _, r := range result {
			So(r.Namespace(), ShouldBeEmpty)
			So(r.Data(), ShouldBeNil)
		}
	})

	Convey("Given valid static metric namespace collect metrics", t, func() {
		So(func() { iostat.CollectMetrics(staticMockMts) }, ShouldNotPanic)
		result, err := iostat.CollectMetrics(staticMockMts)
		So(len(result), ShouldEqual, 6)
		So(err, ShouldBeNil)
		So(result[0].Data(), ShouldEqual, 0.50)
		So(result[1].Data(), ShouldEqual, 0)
		So(result[2].Data(), ShouldEqual, 0.13)
		So(result[3].Data(), ShouldEqual, 0)
		So(result[4].Data(), ShouldEqual, 0)
		So(result[5].Data(), ShouldEqual, 99.37)
	})

	Convey("Given valid dynamic metric namespace collect metrics", t, func() {
		So(func() { iostat.CollectMetrics(dynamicMockMts) }, ShouldNotPanic)
		result, err := iostat.CollectMetrics(dynamicMockMts)
		So(len(result), ShouldEqual, 72)
		So(err, ShouldBeNil)

		m := make(map[string]interface{}, len(result))
		for _, r := range result {
			m[r.Namespace().String()] = r.Data()
		}

		So(len(m), ShouldEqual, len(refMap))
		for key, val := range refMap {
			v, ok := m[key]
			So(ok, ShouldBeTrue)
			So(v, ShouldEqual, val)
		}

		for _, r := range result {
			_, ok := r.Data_.(float64)
			So(ok, ShouldBeTrue)
		}
	})

	Convey("Get metric types", t, func() {
		mts, err := iostat.GetMetricTypes(plugin.ConfigType{})
		So(err, ShouldBeNil)
		So(len(mts), ShouldEqual, 19)

		namespaces := []string{}
		for _, m := range mts {
			namespaces = append(namespaces, m.Namespace().String())
		}

		So(namespaces, ShouldContain, "/intel/iostat/avg-cpu/%idle")
		So(namespaces, ShouldContain, "/intel/iostat/avg-cpu/%iowait")
		So(namespaces, ShouldContain, "/intel/iostat/avg-cpu/%nice")
		So(namespaces, ShouldContain, "/intel/iostat/avg-cpu/%steal")
		So(namespaces, ShouldContain, "/intel/iostat/avg-cpu/%system")
		So(namespaces, ShouldContain, "/intel/iostat/avg-cpu/%user")
		So(namespaces, ShouldContain, "/intel/iostat/device/*/%util")
		So(namespaces, ShouldContain, "/intel/iostat/device/*/avgqu-sz")
		So(namespaces, ShouldContain, "/intel/iostat/device/*/avgrq-sz")
		So(namespaces, ShouldContain, "/intel/iostat/device/*/await")
		So(namespaces, ShouldContain, "/intel/iostat/device/*/r_await")
		So(namespaces, ShouldContain, "/intel/iostat/device/*/r_per_sec")
		So(namespaces, ShouldContain, "/intel/iostat/device/*/rkB_per_sec")
		So(namespaces, ShouldContain, "/intel/iostat/device/*/rrqm_per_sec")
		So(namespaces, ShouldContain, "/intel/iostat/device/*/svctm")
		So(namespaces, ShouldContain, "/intel/iostat/device/*/w_await")
		So(namespaces, ShouldContain, "/intel/iostat/device/*/w_per_sec")
		So(namespaces, ShouldContain, "/intel/iostat/device/*/wkB_per_sec")
		So(namespaces, ShouldContain, "/intel/iostat/device/*/wrqm_per_sec")
	})

	Convey("Get config policy", t, func() {
		policy, err := iostat.GetConfigPolicy()
		So(err, ShouldBeNil)
		So(policy, ShouldResemble, cpolicy.New())
	})
}
