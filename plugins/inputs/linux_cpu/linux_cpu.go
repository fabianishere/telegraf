// +build linux

package linux_cpu

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const defaultHostSys = "/sys"

type LinuxCPU struct {
	Log             telegraf.Logger `toml:"-"`
	PathSysfs       string          `toml:"host_sys"`
	GatherCPUFreq   bool            `toml:"gather_cpufreq"`
	GatherThrottles bool            `toml:"gather_throttles"`
	cpus            []cpu
}

type cpu struct {
	id    string
	path  string
	props map[string]*os.File
}

type prop struct {
	name     string
	path     string
	optional bool
}

var sampleConfig = `
  ## Path for sysfs filesystem.
  ## See https://www.kernel.org/doc/Documentation/filesystems/sysfs.txt
  ## Defaults:
  # host_sys = "/sys"

  ## Gather CPU frequency information
  ## Defaults:
  # gather_cpufreq = true

  ## Gather CPU throttle information per core
  ## Defaults:
  # gather_throttles = false
`

func (g *LinuxCPU) SampleConfig() string {
	return sampleConfig
}

func (g *LinuxCPU) Description() string {
	return "Collects CPU metrics exposed on Linux"
}

func (g *LinuxCPU) Init() error {
	if g.PathSysfs == "" {
		g.PathSysfs = defaultHostSys
	}

	cpus, err := g.discoverCpus()
	if err != nil {
		return err
	}
	g.cpus = cpus

	return nil
}

func (g *LinuxCPU) Gather(acc telegraf.Accumulator) error {
	for _, cpu := range g.cpus {
		fields := make(map[string]interface{})
		tags := map[string]string{"cpu": cpu.id}

		failed := false
		for name, fd := range cpu.props {
			v, err := readUintFromFile(fd)
			if err != nil {
				acc.AddError(err)
				failed = true
				break
			}

			fields[name] = v
		}

		if !failed {
			acc.AddFields("linux_cpu", fields, tags)
		}
	}

	return nil
}

func (g *LinuxCPU) discoverCpus() ([]cpu, error) {
	var cpus []cpu

	glob := path.Join(g.PathSysfs, "devices/system/cpu/cpu[0-9]*")
	cpuDirs, err := filepath.Glob(glob)
	if err != nil {
		return nil, err
	}

	if len(cpuDirs) == 0 {
		return nil, fmt.Errorf("no CPUs detected at: %s", glob)
	}

	for _, dir := range cpuDirs {
		_, cpuName := filepath.Split(dir)
		cpuNum := strings.TrimPrefix(cpuName, "cpu")

		cpu := cpu{
			id:    cpuNum,
			path:  dir,
			props: make(map[string]*os.File),
		}

		var props []prop

		if g.GatherCPUFreq {
			props = append(props,
				prop{ name: "scaling_cur_freq", path: "cpufreq/scaling_cur_freq", optional: false },
				prop{ name: "scaling_min_freq", path: "cpufreq/scaling_min_freq", optional: false },
				prop{ name: "scaling_max_freq", path: "cpufreq/scaling_max_freq", optional: false },
				prop{ name: "cpuinfo_cur_freq", path: "cpufreq/cpuinfo_cur_freq", optional: true },
				prop{ name: "cpuinfo_min_freq", path: "cpufreq/cpuinfo_min_freq", optional: true },
				prop{ name: "cpuinfo_max_freq", path: "cpufreq/cpuinfo_max_freq", optional: true },
			)
		}

		if g.GatherThrottles {
			props = append(
				props,
				prop{ name: "throttle_count", path: "thermal_throttle/core_throttle_count", optional: false },
				prop{ name: "throttle_max_time", path: "thermal_throttle/core_throttle_max_time_ms", optional: false },
				prop{ name: "throttle_total_time", path: "thermal_throttle/core_throttle_total_time_ms", optional: false },
			)
		}

		var failed = false
		for _, prop := range props {
			fd, err := os.Open(filepath.Join(dir, prop.path))
			if err != nil && !prop.optional {
				g.Log.Warnf("Failed to load property %s: %s", filepath.Join(dir, prop.path), err)
				failed = true
				break
			}

			cpu.props[prop.name] = fd
		}

		if len(cpu.props) == 0 {
			g.Log.Warnf("No properties enabled for CPU %s", cpuNum)
			failed = true
		}

		if !failed {
			cpus = append(cpus, cpu)
		}
	}
	return cpus, nil
}

func init() {
	inputs.Add("linux_cpu", func() telegraf.Input {
		return &LinuxCPU{
			GatherCPUFreq: true,
		}
	})
}

func readUintFromFile(fd *os.File) (uint64, error) {
	buffer := make([]byte, 22)

	n, err := fd.ReadAt(buffer, 0)
	if err != nil && err != io.EOF {
		return 0, fmt.Errorf("error on reading file, err: %v", err)
	}

	return strconv.ParseUint(string(buffer[:n-1]), 10, 64)
}
