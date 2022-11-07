DISCONTINUATION OF PROJECT. 

This project will no longer be maintained by Intel.

This project has been identified as having known security escapes.

Intel has ceased development and contributions including, but not limited to, maintenance, bug fixes, new releases, or updates, to this project.  

Intel no longer accepts patches to this project.
# DISCONTINUATION OF PROJECT 

**This project will no longer be maintained by Intel.  Intel will not provide or guarantee development of or support for this project, including but not limited to, maintenance, bug fixes, new releases or updates.  Patches to this project are no longer accepted by Intel. If you have an ongoing need to use this project, are interested in independently developing it, or would like to maintain patches for the community, please create your own fork of the project.**


# Snap collector plugin - iostat

This plugin collects CPU statistics and device I/O statistics based on iostat command tool which is available in every distribution of Linux.

Used for monitoring system input/output device loading by observing the time the devices are active in relation to their average transfer rates. 

The intention is that data will be collected, aggregated and fed into graphing and analysis plugin that can be used to change system configuration to better balance the input/output load between physical disks.

This plugin is used in the [Snap framework](http://github.com/intelsdi-x/snap).


1. [Getting Started](#getting-started)
   * [System Requirements](#system-requirements)
   * [Installation](#installation)
   * [Configuration and Usage](#configuration-and-usage)
2. [Documentation](#documentation)
   * [Collected Metrics](#collected-metrics)
   * [Examples](#examples)
   * [Known limitations](#known-limitations)
   * [Roadmap](#roadmap)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
5. [License](#license)
6. [Acknowledgements](#acknowledgements)

## Getting Started

In order to use this plugin you need "iostat" to be installed on a Linux target host.

### System Requirements

* Linux OS
* [sysstat package](#installation) in version 10.2 or greater
* [golang 1.6+](https://golang.org/dl/)  (needed only for building)

The iostat command-line tool is part of the sysstat package available under the GNU General Public License.

### Installation

#### Install sysstat package:
To install sysstat package from the official repositories simply use:
- For Ubuntu, Debian: `sudo apt-get install sysstat`
- For CentOS, Fedora: `sudo yum install sysstat`

#### Download the plugin binary:

You can get the pre-built binaries for your OS and architecture from the plugin's [GitHub Releases](https://github.com/intelsdi-x/snap-plugin-collector-iostat/releases) page. Download the plugin from the latest release and load it into `snapteld` (`/opt/snap/plugins` is the default location for Snap packages).

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

* [linux iostat tool](http://linux.die.net/man/1/iostat)
* [Snap iostat unit test](iostat/iostat_test.go)
* [Snap iostat examples](#examples)

### Collected Metrics
This plugin has the ability to gather **cpu and device statistics**.   
List of collected metrics is described in [METRICS.md](METRICS.md).


**Notes:** If would like the results since boot you can set the config option `ReportSinceBoot` to `true` (see the sample task below)

### Examples
Example running  iostat collector and writing data to file.

Ensure [Snap daemon is running](https://github.com/intelsdi-x/snap#running-snap):
* initd: `service snap-telemetry start`
* systemd: `systemctl start snap-telemetry`
* command line: `snapteld -l 1 -t 0 &`

Download and load Snap plugins:
```
$ wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-collector-iostat/latest/linux/x86_64/snap-plugin-collector-iostat
$ wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-publisher-file/latest/linux/x86_64/snap-plugin-publisher-file
$ chmod 755 snap-plugin-*
$ snaptel plugin load snap-plugin-collector-iostat
$ snaptel plugin load snap-plugin-publisher-file
```

See all available metrics:

```
$ snaptel metric list
```

Download an [example task file](examples/tasks/iostat-file.json) and load it:
```
$ curl -sfLO https://raw.githubusercontent.com/intelsdi-x/snap-plugin-collector-iostat/master/examples/tasks/iostat-file.json
$ snaptel task create -t iostat-file.json
Using task manifest to create task
Task created
ID: 02dd7ff4-8106-47e9-8b86-70067cd0a850
Name: Task-02dd7ff4-8106-47e9-8b86-70067cd0a850
State: Running
```
This data is published to a file `/tmp/published_iostat` per task specification

See realtime output from `snaptel task watch <task_id>` (CTRL+C to exit)
```
$ snaptel task watch 02dd7ff4-8106-47e9-8b86-70067cd0a850

Watching Task (02dd7ff4-8106-47e9-8b86-70067cd0a850):
NAMESPACE                                        DATA    TIMESTAMP                               SOURCE
/intel/iostat/avg-cpu/%idle                97.62   2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/avg-cpu/%iowait              1.13    2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/avg-cpu/%nice                0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/avg-cpu/%steal               0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/avg-cpu/%system              0.5     2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/avg-cpu/%user                0.75    2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/ALL/%util             1.69    2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/ALL/r_per_sec         0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/ALL/rkB_per_sec       0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/ALL/rrqm_per_sec      0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/ALL/w_per_sec         133     2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/ALL/wkB_per_sec       50672   2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/ALL/wrqm_per_sec      326     2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sda1/%util            0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sda1/r_per_sec        0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sda1/rkB_per_sec      0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sda1/rrqm_per_sec     0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sda1/w_per_sec        0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sda1/wkB_per_sec      0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sda1/wrqm_per_sec     0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb/%util             6.8     2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb/r_per_sec         0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb/rkB_per_sec       0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb/rrqm_per_sec      0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb/w_per_sec         68      2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb/wkB_per_sec       25336   2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb/wrqm_per_sec      163     2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb1/%util            6.7     2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb1/r_per_sec        0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb1/rkB_per_sec      0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb1/rrqm_per_sec     0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb1/w_per_sec        65      2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb1/wkB_per_sec      25336   2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb1/wrqm_per_sec     163     2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb2/%util            0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb2/r_per_sec        0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb2/rkB_per_sec      0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb2/rrqm_per_sec     0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb2/w_per_sec        0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb2/wkB_per_sec      0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
/intel/iostat/device/sdb2/wrqm_per_sec     0       2015-12-01 05:39:57.79589855 -0500 EST  gklab-108-109-110-111
```

Stop task:
```
$ snaptel task stop 02dd7ff4-8106-47e9-8b86-70067cd0a850
Task stopped:
ID: 02dd7ff4-8106-47e9-8b86-70067cd0a850
```

### Known limitations
Expectation of numeric string with a dot as decimal separator.

For another decimal mark, the following error will occur: 
```
Common error: invalid metric value strconv.ParseFloat: parsing "0,00": invalid syntax
```
To resolve that, the locale numeric configuration (LC_NUMERIC) needs to be changed to set dot as decimal separator.


### Roadmap
As we launch this plugin, we do not have any outstanding requirements for the next release. If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-collector-iostat/issues) 
and/or submit a [pull request](https://github.com/intelsdi-x/snap-plugin-collector-iostat/pulls).

## Community Support
This repository is one of **many** plugins in **Snap**, a powerful telemetry framework. See the full project at http://github.com/intelsdi-x/snap 

To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support) or visit [Slack](http://slack.snap-telemetry.io).

## Contributing
We love contributions!

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[Snap](http://github.com/intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).


## Acknowledgements

* Author: [Izabella Raulin](https://github.com/IzabellaRaulin)

And **thank you!** Your contribution, through code and participation, is incredibly important to us.
