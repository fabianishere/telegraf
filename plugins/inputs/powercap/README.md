# Powercap Input Plugin

The `powercap` plugin gathers metrics from Linux' power capping framework.

#### Configuration
```toml
[[inputs.powercap]]
  ## Path for sysfs filesystem.
  ## See https://www.kernel.org/doc/Documentation/filesystems/sysfs.txt
  ## Defaults:
  # host_sys = "/sys"
  
  ## Control type to monitor
  ## Defaults:
  control_type = "intel-rapl"
```

### Metrics
- powercap

  - The following tags are emitted by the plugin:

    | Tag | Description |
    |-----|-------------|
    | `zone_id` | Identifier of the power zone |
    | `zone_name` | Name of the power zone | 

  - The following fields are emitted by the plugin:

    | Metric name (field) | Description | Units |
    |-----|-------------|-----|
    | `enabled` | Flag to indicate control at zone has been enabled/disabled |  |
    | `energy_uj` | Current energy counter | Microjoules (µJ) |
    | `power_uw` | Average power over collection interval | Microwatts (µW) |

### Example Output
```
powercap,host=go,zone_id=1,zone_name=psys enabled=false,power_uw=249443932i 1615671430000000000
powercap,host=go,zone_id=1,zone_name=psys energy_uj=83348191596i 1615671430000000000
powercap,host=go,zone_id=0,zone_name=package-0 power_uw=75171313i,enabled=true 1615671430000000000
powercap,host=go,zone_id=0,zone_name=package-0 energy_uj=183846238275i 1615671430000000000
powercap,host=go,zone_id=0:0,zone_name=core power_uw=11533257i,enabled=false 1615671430000000000
powercap,host=go,zone_id=0:0,zone_name=core energy_uj=203042442551i 1615671430000000000
powercap,host=go,zone_id=0:1,zone_name=uncore enabled=false,power_uw=0i 1615671430000000000
powercap,host=go,zone_id=0:1,zone_name=uncore energy_uj=90148i 1615671430000000000
powercap,host=go,zone_id=0:2,zone_name=dram enabled=false,power_uw=28943184i 1615671430000000000
powercap,host=go,zone_id=0:2,zone_name=dram energy_uj=13845530290i 1615671430000000000
```