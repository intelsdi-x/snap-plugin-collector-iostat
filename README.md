# Snap collector plugin - iostat

This plugin collects CPU statistics and device I/O statistics based on iostat command tool which is available in every distribution of Linux.

Used for monitoring system input/output device loading by observing the time the devices are active in relation to their average transfer rates. 

The intention is that data will be collected, aggregated and fed into graphing and analysis plugin that can be used to change system configuration to better balance the input/output load between physical disks.

This plugin is used in the [Snap framework] (http://github.com/intelsdi-x/snap).


1. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Installation](#installation)
  * [Configuration and Usage](#configuration-and-usage)
2. [Documentation](#documentation)
  * [Collected Metrics](#collected-metrics)
  * [Examples](#examples)
  * [Roadmap](#roadmap)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
5. [License](#license)
6. [Acknowledgements](#acknowledgements)

## Getting Started

In order to use this plugin you need "iostat" to be installed on a Linux target host.

### System Requirements

* Linux OS
* [sysstat package] (#installation)
 * sysstat 10.2 or greater
* [golang 1.5+](https://golang.org/dl/)  (needed only for building)

The iostat command-line tool is part of the sysstat package available under the GNU General Public License.

### Installation

#### Install sysstat package:
To install sysstat package from the official repositories simply use:
- For Ubuntu, Debian: `sudo apt-get install sysstat`
- For CentOS, Fedora: `sudo yum install sysstat`

#### Download the plugin binary:

You can get the pre-built binaries for your OS and architecture from the plugin's [GitHub Releases](https://github.com/intelsdi-x/snap-plugin-collector-iostat/releases) page. Download the plugin from the latest release and load it into `snapd` (`/opt/snap/plugins` is the default location for Snap packages).

#### To build the plugin binary:

Fork https://github.com/intelsdi-x/snap-plugin-collector-iostat
Clone repo into `$GOPATH/src/github.com/intelsdi-x/`:

```
$ git clone https://github.com/<yourGithubID>/snap-plugin-collector-iostat.git
```

Build the Snap iostat plugin by running make within the cloned repo:
```
$ make
```
This builds the plugin in `./build/`

### Configuration and Usage
* Set up the [Snap framework](https://github.com/intelsdi-x/snap#getting-started)
* Load the plugin and create a task, see example in [Examples](#examples).

By default iostat executable binary are searched in the directories named by the PATH environment. 
Customize path to iostat executable is also possible by setting environment variable `export SNAP_IOSTAT_PATH=/path/to/iostat/bin`

## Documentation

To learn more about this plugin and iostat tool, visit:

* [linux iostat tool] (http://linux.die.net/man/1/iostat)
* [Snap iostat unit test](iostat/iostat_test.go)
* [Snap iostat examples](#examples)

### Collected Metrics
This plugin has the ability to gather the following metrics:

* **CPU statistic**

Metric namespace prefix: /intel/linux/iostat/avg-cpu/

Namespace | Description 
------------ | -------------
%user | The percentage of CPU utilization that occurred while executing at the user level (the application usage)
%nice | The percentage of CPU utilization that occurred while executing at the user level with nice priority
%system | The percentage of CPU utilization that occurred while executing at the system level (the kernel usage)
%iowait | The percentage of time that the CPU or CPUs were idle during which the system had an outstanding disk I/O request
%steal | The percentage of time spent in involuntary wait by the virtual CPU or CPUs while the hypervisor was servicing another virtual processor
%idle | The percentage of time that the CPU or CPUs were idle and the systems did not have an outstanding disk I/O request



* **Device statistics**

Metric namespace prefix: /intel/linux/iostat/device/{disk_or_partition}

Name | Description 
------------ | -------------
rrqm_per_sec | The number of read requests merged per second queued to the device
wrqm_per_sec |The number of write requests merged per second queued to the device
r_per_sec | The number of read requests issued to the device per second
w_per_sec | The number of write requests issued to the device per second
rkB_per_sec | The number of kilobytes read from the device per second
wkB_per_sec | The number of kilobytes written to the device per second
avgrq-sz | The average size (in sectors) of the requests issued to the device
avgqu-sz | The average queue length of the requests issued to the device
await | The average time (milliseconds) for I/O requests issued to the device to be served This includes the time spent by the requests in queue and the time spent servicing them
r_await | The average time (in milliseconds) for read requests issued to the device to be served which includes the time spent by the requests in queue and the time spent servicing them
w_await | The average time (in milliseconds) for write requests issued to the device to be served which includes the time spent by the requests in queue and the time spent servicing them
svctm | The average service time (in milliseconds) for I/O requests issued to the device - Warning! Do not trust this field; it will be removed in a future version of sysstat
%util | Percentage of CPU time during which I/O requests were issued to the device (bandwidth utilization for the device); device saturation occurs when this values is close to 100%


*Notes:*

* The total number of read and write requests issued to the device per second equals the number of transaction per second	
 * tps=r_per_sec+w_per_sec
* The metrics are sampled over 1 second   
 * If would like the results since boot you can set the config option `ReportSinceBoot` to `true` (see the sample task below)

### Examples
Example running  iostat collector and writing data to file.

Ensure [Snap daemon is running](https://github.com/intelsdi-x/snap#running-snap):
* initd: `service snap-telemetry start`
* systemd: `systemctl start snap-telemetry`
* command line: `snapd -l 1 -t 0 &`

Download and load Snap plugins:
```
$ wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-collector-iostat/latest/linux/x86_64/snap-plugin-collector-iostat
$ wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-publisher-file/latest/linux/x86_64/snap-plugin-publisher-file
$ chmod 755 snap-plugin-*
$ snapctl plugin load snap-plugin-collector-iostat
$ snapctl plugin load snap-plugin-publisher-file
```

See all available metrics:

```
$ snapctl metric list
```

Download an [example task file](examples/tasks/iostat-file.json) and load it:
```
$ curl -sfLO https://raw.githubusercontent.com/intelsdi-x/snap-plugin-collector-iostat/master/examples/tasks/iostat-file.json
$ snapctl task create -t iostat-file.json
Using task manifest to create task
Task created
ID: 02dd7ff4-8106-47e9-8b86-70067cd0a850
Name: Task-02dd7ff4-8106-47e9-8b86-70067cd0a850
State: Running
```
This data is published to a file `/tmp/published_iostat` per task specification

See realtime output from `snapctl task watch <task_id>` (CTRL+C to exit)
```
$ snapctl task watch 02dd7ff4-8106-47e9-8b86-70067cd0a850

Watching Task (02dd7ff4-8106-47e9-8b86-70067cd0a850):
NAMESPACE                                        DATA    TIMESTAMP                               SOURCE
/intel/linux/iostat/avg-cpu/%idle                97.62   2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/avg-cpu/%iowait              1.13    2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/avg-cpu/%nice                0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/avg-cpu/%steal               0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/avg-cpu/%system              0.5     2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/avg-cpu/%user                0.75    2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/ALL/%util             1.69    2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/ALL/r_per_sec         0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/ALL/rkB_per_sec       0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/ALL/rrqm_per_sec      0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/ALL/w_per_sec         133     2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/ALL/wkB_per_sec       50672   2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/ALL/wrqm_per_sec      326     2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sda1/%util            0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sda1/r_per_sec        0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sda1/rkB_per_sec      0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sda1/rrqm_per_sec     0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sda1/w_per_sec        0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sda1/wkB_per_sec      0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sda1/wrqm_per_sec     0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb/%util             6.8     2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb/r_per_sec         0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb/rkB_per_sec       0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb/rrqm_per_sec      0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb/w_per_sec         68      2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb/wkB_per_sec       25336   2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb/wrqm_per_sec      163     2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb1/%util            6.7     2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb1/r_per_sec        0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb1/rkB_per_sec      0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb1/rrqm_per_sec     0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb1/w_per_sec        65      2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb1/wkB_per_sec      25336   2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb1/wrqm_per_sec     163     2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb2/%util            0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb2/r_per_sec        0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb2/rkB_per_sec      0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb2/rrqm_per_sec     0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb2/w_per_sec        0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb2/wkB_per_sec      0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/linux/iostat/device/sdb2/wrqm_per_sec     0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
```

Stop task:
```
$ $SNAP_PATH/bin/snapctl task stop 02dd7ff4-8106-47e9-8b86-70067cd0a850
Task stopped:
ID: 02dd7ff4-8106-47e9-8b86-70067cd0a850
```

### Roadmap
As we launch this plugin, we do not have any outstanding requirements for the next release. If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-collector-iostat/issues) 
and/or submit a [pull request](https://github.com/intelsdi-x/snap-plugin-collector-iostat/pulls).

## Community Support
This repository is one of **many** plugins in **Snap**, a powerful telemetry framework. See the full project at http://github.com/intelsdi-x/snap.

To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support).

## Contributing
We love contributions!

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[Snap](http://github.com/intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).


## Acknowledgements

* Author: [Izabella Raulin](https://github.com/IzabellaRaulin)

And **thank you!** Your contribution, through code and participation, is incredibly important to us.
