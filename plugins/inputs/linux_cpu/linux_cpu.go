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
	GatherThrottles bool            `toml:"gather_throttles"`
	cpus            []cpu
}

type cpu struct {
	id    string
	path  string
	props map[string]*os.File
}

var sampleConfig = `
  ## Path for sysfs filesystem.
  ## See https://www.kernel.org/doc/Documentation/filesystems/sysfs.txt
  ## Defaults:
  # host_sys = "/sys"

  ## Gather CPU throttles per core
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

		if _, err := os.Stat(filepath.Join(dir, "cpufreq")); os.IsNotExist(err) {
			return nil, fmt.Errorf("CPUFreq subsystem not available for cpu %s", cpuNum)
		}

		cpu := cpu{
			id:    cpuNum,
			path:  dir,
			props: make(map[string]*os.File),
		}

		props := make(map[string]string)
		props["cpufreq/scaling_cur_freq"] = "scaling_cur_freq"
		props["cpufreq/scaling_min_freq"] = "scaling_min_freq"
		props["cpufreq/scaling_max_freq"] = "scaling_max_freq"

		if g.GatherThrottles {
			props["thermal_throttle/core_throttle_count"] = "throttle_count"
		}

		var failed = false
		for prop, name := range props {
			fd, err := os.Open(filepath.Join(dir, prop))
			if err != nil {
				g.Log.Warnf("Failed to load property %s: %s", filepath.Join(dir, prop), err)
				failed = true
				break
			}

			cpu.props[name] = fd
		}

		if !failed {
			cpus = append(cpus, cpu)
		}
	}
	return cpus, nil
}

func init() {
	inputs.Add("linux_cpu", func() telegraf.Input { return &LinuxCPU{} })
}

func readUintFromFile(fd *os.File) (uint64, error) {
	buffer := make([]byte, 22)

	n, err := fd.ReadAt(buffer, 0)
	if err != nil && err != io.EOF {
		return 0, fmt.Errorf("error on reading file, err: %v", err)
	}

	return strconv.ParseUint(string(buffer[:n-1]), 10, 64)
}
