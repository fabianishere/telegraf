# Linux CPU Input Plugin
The `linux_cpu` plugin gathers CPU metrics exposed on Linux-based systems.

#### Configuration
```toml
# Collects current CPU's frequencies
[[inputs.cpufreq]]
  ## Path for sysfs filesystem.
  ## See https://www.kernel.org/doc/Documentation/filesystems/sysfs.txt
  ## Defaults:
  # host_sys = "/sys"
 
  ## Gather CPU frequency information
  ## Defaults:
  # gather_cpufreq = true

  ## Gather CPU throttles per core
  ## Defaults:
  # gather_throttles = false
```

### Metrics
- linux_cpu

  - The following tags are emitted by the plugin:

    | Tag | Description |
    |-----|-------------|
    | `cpu` | Identifier of the CPU |

  - The following fields are emitted by the plugin when enabling `gather_cpufreq`:

    | Metric name (field) | Description | Units |
    |---------------------|-------------|-------|
    | `scaling_cur_freq` | Current frequency of the CPU as determined by CPUFreq | KHz |
    | `scaling_min_freq` | Minimum frequency the governor can scale to | KHz |
    | `scaling_max_freq` | Maximum frequency the governor can scale to | KHz |
    | `cpuinfo_cur_freq` | Current frequency of the CPU as determined by the hardware | KHz |
    | `cpuinfo_min_freq` | Minimum operating frequency of the CPU | KHz |
    | `cpuinfo_max_freq` | Maximum operating frequency of the CPU | KHz |

  - The following fields are emitted by the plugin when enabling `gather_throttles`:

    | Metric name (field) | Description | Units |
    |---------------------|-------------|-------|
    | `throttle_count` | Number of thermal throttle events reported by the CPU |  |
    | `throttle_max_time` | Maximum amount of time CPU was in throttled state  | ms |
    | `throtlle_total_time` | Cumulative time during which the CPU was in throttled state | ms |


### Example Output

```
> linux_cpu,cpu=0,host=Z370M-DS3H scaling_cur_freq=1382522i,scaling_max_freq=4900000i,scaling_min_freq=800000i 1604049750000000000
> linux_cpu,cpu=1,host=Z370M-DS3H scaling_cur_freq=1094884i,scaling_max_freq=4900000i,scaling_min_freq=800000i 1604049750000000000
> linux_cpu,cpu=2,host=Z370M-DS3H scaling_cur_freq=1010482i,scaling_max_freq=4900000i,scaling_min_freq=800000i 1604049750000000000
> linux_cpu,cpu=3,host=Z370M-DS3H scaling_cur_freq=2089249i,scaling_max_freq=4900000i,scaling_min_freq=800000i 1604049750000000000
> linux_cpu,cpu=4,host=Z370M-DS3H scaling_cur_freq=1272475i,scaling_max_freq=4900000i,scaling_min_freq=800000i 1604049750000000000
> linux_cpu,cpu=5,host=Z370M-DS3H scaling_cur_freq=1374903i,scaling_max_freq=4900000i,scaling_min_freq=800000i 1604049750000000000
> linux_cpu,cpu=6,host=Z370M-DS3H scaling_cur_freq=1355753i,scaling_max_freq=4900000i,scaling_min_freq=800000i 1604049750000000000
> linux_cpu,cpu=7,host=Z370M-DS3H scaling_cur_freq=1153656i,scaling_max_freq=4900000i,scaling_min_freq=800000i 1604049750000000000
```
