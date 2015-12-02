# snap collector plugin - iostat

This plugin collects CPU statistics and device I/O statistics based on iostat command tool which is available in every distribution of Linux.

Used for monitoring system input/output device loading by observing the time the devices are active in relation to their average transfer rates. 

The intention is that data will be collected, aggregated and fed into graphing and analysis plugin that can be used to change system configuration to better balance the input/output load between physical disks.

This plugin is used in the [snap framework] (http://github.com/intelsdi-x/snap).


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
* [golang 1.4+](https://golang.org/dl/)

The iostat command-line tool is part of the sysstat package available under the GNU General Public License.

### Installation

#### Install sysstat package:
To install sysstat package from the official repositories simply use:
- For Ubuntu, Debian: `sudo apt-get install sysstat`
- For CentOS, Fedora: `sudo yum install sysstat`

#### To build the plugin binary:
Fork https://github.com/intelsdi-x/snap-plugin-collector-iostat  
Clone repo into `$GOPATH/src/github.com/intelsdi-x/`:

```
$ git clone https://github.com/<yourGithubID>/snap-plugin-collector-iostat.git
```

Build the plugin by running make within the cloned repo:
```
$ make
```
This builds the plugin in `/build/rootfs/`

### Configuration and Usage
* Set up the [snap framework](https://github.com/intelsdi-x/snap/blob/master/README.md#getting-started)
* Ensure `$SNAP_PATH` is exported  
`export SNAP_PATH=$GOPATH/src/github.com/intelsdi-x/snap/build`

By default iostat executable binary are searched in the directories named by the PATH environment. 
Customize path to iostat executable is also possible by setting environment variable `export SNAP_IOSTAT_PATH=/path/to/iostat/bin`

## Documentation

To learn more about this plugin and iostat tool, visit:

* [linux iostat tool] (http://linux.die.net/man/1/iostat)
* [snap iostat unit test](https://github.com/intelsdi-x/snap-plugin-collector-iostat/blob/master/iostat/iostat_test.go)
* [snap iostat examples](#examples)

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

The total number of read and write requests issued to the device per second equals the number of transaction per second:	tps=r_per_sec+w_per_sec

By default metrics are gathered once per second.

### Examples
Example running  iostat collector and writing data to file.

In one terminal window, open the snap daemon:
```
$ snapd -l 1 -t 0
```

In another terminal window, load iostat plugin for collecting:
```
$ snapctl plugin load $SNAP_IOSTAT_PLUGIN_DIR/build/rootfs/snap-plugin-collector-iostat
Plugin loaded
Name: iostat
Version: 1
Type: collector
Signed: false
Loaded Time: Tue, 01 Dec 2015 05:19:48 EST
```
See available metrics for your system:
```
$ snapctl metric list
```

Load file plugin for publishing:
```
$ snapctl plugin load $SNAP_DIR/build/plugin/snap-publisher-file
Plugin loaded
Name: file
Version: 3
Type: publisher
Signed: false
Loaded Time: Tue, 01 Dec 2015 05:21:18 EST
```

Create a task JSON file (exemplary file in examples/tasks/iostat-file.json):  
```json
{
    "version": 1,
    "schedule": {
        "type": "simple",
        "interval": "1s"
    },
    "workflow": {
        "collect": {
            "metrics": {
                "/intel/linux/iostat/avg-cpu/%user": {},
                "/intel/linux/iostat/avg-cpu/%system": {},
                "/intel/linux/iostat/avg-cpu/%idle": {},
				"/intel/linux/iostat/avg-cpu/%iowait": {},
				"/intel/linux/iostat/avg-cpu/%nice": {},
				"/intel/linux/iostat/avg-cpu/%steal": {},
                "/intel/linux/iostat/device/sda1/rrqm_per_sec": {},
                "/intel/linux/iostat/device/sda1/wrqm_per_sec": {},
                "/intel/linux/iostat/device/sda1/r_per_sec": {},
                "/intel/linux/iostat/device/sda1/w_per_sec": {},
                "/intel/linux/iostat/device/sda1/rkB_per_sec": {},
                "/intel/linux/iostat/device/sda1/wkB_per_sec": {},
                "/intel/linux/iostat/device/sda1/%util": {},
                "/intel/linux/iostat/device/sdb/rrqm_per_sec": {},
                "/intel/linux/iostat/device/sdb/wrqm_per_sec": {},
                "/intel/linux/iostat/device/sdb/r_per_sec": {},
                "/intel/linux/iostat/device/sdb/w_per_sec": {},
                "/intel/linux/iostat/device/sdb/rkB_per_sec": {},
                "/intel/linux/iostat/device/sdb/wkB_per_sec": {},
                "/intel/linux/iostat/device/sdb/%util": {},
                "/intel/linux/iostat/device/sdb1/rrqm_per_sec": {},
                "/intel/linux/iostat/device/sdb1/wrqm_per_sec": {},
                "/intel/linux/iostat/device/sdb1/r_per_sec": {},
                "/intel/linux/iostat/device/sdb1/w_per_sec": {},
                "/intel/linux/iostat/device/sdb1/rkB_per_sec": {},
                "/intel/linux/iostat/device/sdb1/wkB_per_sec": {},
                "/intel/linux/iostat/device/sdb1/%util": {},
                "/intel/linux/iostat/device/sdb2/rrqm_per_sec": {},
                "/intel/linux/iostat/device/sdb2/wrqm_per_sec": {},
                "/intel/linux/iostat/device/sdb2/r_per_sec": {},
                "/intel/linux/iostat/device/sdb2/w_per_sec": {},
                "/intel/linux/iostat/device/sdb2/rkB_per_sec": {},
                "/intel/linux/iostat/device/sdb2/wkB_per_sec": {},
                "/intel/linux/iostat/device/sdb2/%util": {},
                "/intel/linux/iostat/device/ALL/rrqm_per_sec": {},
                "/intel/linux/iostat/device/ALL/wrqm_per_sec": {},
                "/intel/linux/iostat/device/ALL/r_per_sec": {},
                "/intel/linux/iostat/device/ALL/w_per_sec": {},
                "/intel/linux/iostat/device/ALL/rkB_per_sec": {},
                "/intel/linux/iostat/device/ALL/wkB_per_sec": {},
                "/intel/linux/iostat/device/ALL/%util": {}
            },
            "config": {
                "/intel/linux/iostat": {
                    "user": "root",
                    "password": "secret"
                }
            },
            "process": null,
            "publish": [
                {
                    "plugin_name": "file",
                    "plugin_version": 3,
                    "config": {
                        "file": "/tmp/published_iostat"
                    }
                }
            ]
        }
    }
}
```

Create a task:
```
$ snapctl task create -t $SNAP_IOSTAT_PLUGIN_DIR/examples/tasks/iostat-file.json
Using task manifest to create task
Task created
ID: 02dd7ff4-8106-47e9-8b86-70067cd0a850
Name: Task-02dd7ff4-8106-47e9-8b86-70067cd0a850
State: Running
```

See sample output from `snapctl task watch <task_id>`

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
(Keys `ctrl+c` terminate task watcher)

These data are published to file and stored there (in this example in /tmp/published_iostat).

Stop task:
```
$ $SNAP_PATH/bin/snapctl task stop 02dd7ff4-8106-47e9-8b86-70067cd0a850
Task stopped:
ID: 02dd7ff4-8106-47e9-8b86-70067cd0a850
```

### Roadmap
This plugin is in active development. As we launch this plugin, we have a few items in mind for the next release:
- [ ] Use channels instead "for" loop to execute iostat cmd

If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-collector-iostat/issues) 
and/or submit a [pull request](https://github.com/intelsdi-x/snap-plugin-collector-iostat/pulls).

## Community Support
This repository is one of **many** plugins in the **snap**, a powerful telemetry agent framework. See the full project at 
http://github.com/intelsdi-x/snap. To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support).


## Contributing
We love contributions! :heart_eyes:

There is more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[snap](http://github.com/intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).


## Acknowledgements

* Author: [Izabella Raulin](https://github.com/IzabellaRaulin)

And **thank you!** Your contribution, through code and participation, is incredibly important to us.
