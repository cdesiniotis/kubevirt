<!--
	This is an auto-generated file.
	PLEASE DO NOT EDIT THIS FILE.
	See "Developing new metrics" below how to generate this file
-->

# KubeVirt metrics
This document aims to help users that are not familiar with all metrics exposed by different KubeVirt components.
All metrics documented here are auto-generated by the utility tool `tools/doc-generator` and reflects exactly what is being exposed.

## KubeVirt Metrics List
### kubevirt_info
Version information.

### kubevirt_migrate_vmi_data_processed_bytes
The total Guest OS data processed and migrated to the new VM. Type: Gauge.

### kubevirt_migrate_vmi_data_remaining_bytes
The remaining guest OS data to be migrated to the new VM. Type: Gauge.

### kubevirt_migrate_vmi_dirty_memory_rate_bytes
The rate of memory being dirty in the Guest OS. Type: Gauge.

### kubevirt_migrate_vmi_disk_transfer_rate_bytes
The rate at which the disk is being transferred. Type: Gauge.

### kubevirt_migrate_vmi_memory_transfer_rate_bytes
The rate at which the memory is being transferred. Type: Gauge.

### kubevirt_virt_controller_leading
Indication for an operating virt-controller. Type: Gauge.

### kubevirt_virt_controller_ready
Indication for a virt-controller that is ready to take the lead. Type: Gauge.

### kubevirt_vmi_cpu_affinity
The vcpu affinity details. Type: Counter.

### kubevirt_vmi_filesystem_capacity_bytes_total
Total VM filesystem capacity in bytes. Type: Gauge.

### kubevirt_vmi_filesystem_used_bytes
Used VM filesystem capacity in bytes. Type: Gauge.

### kubevirt_vmi_memory_actual_balloon_bytes
Current balloon bytes. Type: Gauge.

### kubevirt_vmi_memory_available_bytes
Amount of `usable` memory as seen by the domain. Type: Gauge.

### kubevirt_vmi_memory_domain_bytes_total
The amount of memory in bytes allocated to the domain. The `memory` value in domain xml file. Type: Gauge.

### kubevirt_vmi_memory_pgmajfault
The number of page faults when disk IO was required. Type: Counter.

### kubevirt_vmi_memory_pgminfault
The number of other page faults, when disk IO was not required. Type: Counter.

### kubevirt_vmi_memory_resident_bytes
Resident set size of the process running the domain. Type: Gauge.

### kubevirt_vmi_memory_swap_in_traffic_bytes_total
Swap in memory traffic in bytes. Type: Gauge.

### kubevirt_vmi_memory_swap_out_traffic_bytes_total
Swap out memory traffic in bytes. Type: Gauge.

### kubevirt_vmi_memory_unused_bytes
Amount of `unused` memory as seen by the domain. Type: Gauge.

### kubevirt_vmi_memory_usable_bytes
The amount of memory which can be reclaimed by balloon without causing host swapping in bytes. Type: Gauge.

### kubevirt_vmi_memory_used_bytes
Amount of `used` memory as seen by the domain. Type: Gauge.

### kubevirt_vmi_network_receive_bytes_total
Network traffic receive in bytes. Type: Counter.

### kubevirt_vmi_network_receive_errors_total
Network receive error packets. Type: Counter.

### kubevirt_vmi_network_receive_packets_dropped_total
The number of rx packets dropped on vNIC interfaces. Type: Counter.

### kubevirt_vmi_network_receive_packets_total
Network traffic receive packets. Type: Counter.

### kubevirt_vmi_network_traffic_bytes_total
Deprecated. Type: Counter.

### kubevirt_vmi_network_transmit_bytes_total
Network traffic transmit in bytes. Type: Counter.

### kubevirt_vmi_network_transmit_errors_total
Network transmit error packets. Type: Counter.

### kubevirt_vmi_network_transmit_packets_dropped_total
The number of tx packets dropped on vNIC interfaces. Type: Counter.

### kubevirt_vmi_network_transmit_packets_total
Network traffic transmit packets. Type: Counter.

### kubevirt_vmi_non_evictable
Indication for a VirtualMachine that its eviction strategy is set to Live Migration but is not migratable. Type: Gauge.

### kubevirt_vmi_outdated_count
Indication for the number of VirtualMachineInstance workloads that are not running within the most up-to-date version of the virt-launcher environment. Type: Gauge.

### kubevirt_vmi_phase_count
Sum of VMIs per phase and node.

`phase` can be one of the following: [`Pending`, `Scheduling`, `Scheduled`, `Running`, `Succeeded`, `Failed`, `Unknown`]. Type: Gauge.

### kubevirt_vmi_storage_flush_requests_total
Storage flush requests. Type: Counter.

### kubevirt_vmi_storage_flush_times_ms_total
Total time (ms) spent on cache flushing. Type: Counter.

### kubevirt_vmi_storage_iops_read_total
I/O read operations. Type: Counter.

### kubevirt_vmi_storage_iops_write_total
I/O write operations. Type: Counter.

### kubevirt_vmi_storage_read_times_ms_total
Storage read operation time. Type: Counter.

### kubevirt_vmi_storage_read_traffic_bytes_total
Storage read traffic in bytes. Type: Counter.

### kubevirt_vmi_storage_write_times_ms_total
Storage write operation time. Type: Counter.

### kubevirt_vmi_storage_write_traffic_bytes_total
Storage write traffic in bytes. Type: Counter.

### kubevirt_vmi_vcpu_seconds
Amount of time spent in each state by each vcpu. Where `id` is the vcpu identifier and `state` can be one of the following: [`OFFLINE`, `RUNNING`, `BLOCKED`]. Type: Counter.

### kubevirt_vmi_vcpu_wait_seconds
Amount of time spent by each vcpu while waiting on I/O. Type: Counter.

## Developing new metrics
After developing new metrics or changing old ones, please run `make generate` to regenerate this document.

If you feel that the new metric doesn't follow these rules, please change `doc-generator` with your needs.
