# snap collector plugin - iostat

## Collected Metrics

This plugin has the ability to gather the following metrics:

  - **CPU statistics**, represented by the metrics with prefix `/intel/iostat/avg-cpu/`
  - **Device statistics**, represented by the metrics with prefix `/intel/iostat/device/`

Namespace | Data Type | Description 
----------| ----------|------------ 
/intel/iostat/avg-cpu/%user | float64 | The percentage of CPU utilization that occurred while executing at the user level (the application usage) 
/intel/iostat/avg-cpu/%nice | float64 | The percentage of CPU utilization that occurred while executing at the user level with nice priority 
/intel/iostat/avg-cpu/%system | float64 | The percentage of CPU utilization that occurred while executing at the system level (the kernel usage) 
/intel/iostat/avg-cpu/%iowait | float64 | The percentage of time that the CPU or CPUs were idle during which the system had an outstanding disk I/O request 
/intel/iostat/avg-cpu/%steal | float64 | The percentage of time spent in involuntary wait by the virtual CPU or CPUs while the hypervisor was servicing another virtual processor
/intel/iostat/avg-cpu/%idle | float64 | The percentage of time that the CPU or CPUs were idle and the systems did not have an outstanding disk I/O request
/intel/iostat/device/[device_id]/rrqm_per_sec | float64 | The number of read requests merged per second queued to the device
/intel/iostat/device/[device_id]/wrqm_per_sec | float64 | The number of write requests merged per second queued to the device
/intel/iostat/device/[device_id]/r_per_sec | float64 | The number of read requests issued to the device per second
/intel/iostat/device/[device_id]/w_per_sec | float64 | The number of write requests issued to the device per second
/intel/iostat/device/[device_id]/rkB_per_sec | float64 | The number of kilobytes read from the device per second
/intel/iostat/device/[device_id]/wkB_per_sec | float64 | The number of kilobytes written to the device per second
/intel/iostat/device/[device_id]/avgrq-sz | float64 | The average size (in sectors) of the requests issued to the device
/intel/iostat/device/[device_id]/avgqu-sz | float64 | The average queue length of the requests issued to the device
/intel/iostat/device/[device_id]/await | float64 | The average time (milliseconds) for I/O requests issued to the device to be served This includes the time spent by the requests in queue and the time spent servicing them
/intel/iostat/device/[device_id]/r_await | float64 | The average time (in milliseconds) for read requests issued to the device to be served which includes the time spent by the requests in queue and the time spent servicing them
/intel/iostat/device/[device_id]/w_await | float64 | The average time (in milliseconds) for write requests issued to the device to be served which includes the time spent by the requests in queue and the time spent servicing them
/intel/iostat/device/[device_id]/svctm | float64 | The average service time (in milliseconds) for I/O requests issued to the device - Warning! Do not trust this field; it will be removed in a future version of sysstat
/intel/iostat/device/[device_id]/%util | float64 | Percentage of CPU time during which I/O requests were issued to the device (bandwidth utilization for the device); device saturation occurs when this values is close to 100%


*Notes:*

* The total number of read and write requests issued to the device per second equals the number of transaction per second	
   * tps=r_per_sec+w_per_sec
* The metrics are sampled over 1 second   
* If would like the results since boot you can set the config option `ReportSinceBoot` to `true`, see how it is done in an [examplary task manifest](examples/tasks/iostat-file.json#L33)
