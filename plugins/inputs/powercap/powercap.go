// +build linux

package powercap

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHostSys = "/sys"
	defaultControlType = "intel-rapl"
	powercapPath = "devices/virtual/powercap"
)

type Powercap struct {
	Log	            telegraf.Logger `toml:"-"`
	PathSysfs       string `toml:"host_sys"`
	ControlType     string `toml:"control_type"`
	Zones           map[string]*Zone
}

type Zone struct {
	path        string
	id          string
	name        string
	parent      *Zone
	children    map[string]*Zone
	fds         map[string]*os.File
	enabled     bool
	energyUj    uint64
	powerUw     uint64
	tsc         int64
}

var sampleConfig = `
  ## Path for sysfs filesystem.
  ## See https://www.kernel.org/doc/Documentation/filesystems/sysfs.txt
  ## Defaults:
  # host_sys = "/sys"

  ## Control type to monitor
  ## Defaults:
  control_type = "intel-rapl"
`

func (g *Powercap) SampleConfig() string {
	return sampleConfig
}

func (g *Powercap) Description() string {
	return "Collect Linux's powercap metrics"
}

func (g *Powercap) Init() error {
	if g.PathSysfs == "" && os.Getenv("HOST_SYS") != "" {
		g.PathSysfs = os.Getenv("HOST_SYS")
	} else {
		g.PathSysfs = defaultHostSys
	}

	if g.ControlType == "" {
		g.ControlType = defaultControlType
	}

	zones, err := g.DiscoverZones(filepath.Join(g.PathSysfs, powercapPath, g.ControlType), nil)
	if err != nil {
		return err
	}
	g.Zones = zones

	for _, zone := range zones {
		err := UpdateZones(zone)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Powercap) Gather(acc telegraf.Accumulator) error {
	for _, zone := range g.Zones {
		g.GatherZone(acc, zone)
	}

	return nil
}

func (g *Powercap) GatherZone(acc telegraf.Accumulator, zone *Zone) {
	err := zone.Update()

	if err != nil {
		acc.AddError(err)
	} else {
		tags := map[string]string{
			"zone_id": zone.id,
			"zone_name": zone.name,
		}

		acc.AddGauge("powercap", map[string]interface{}{
			"enabled": zone.enabled,
			"power_uw": zone.powerUw,
		}, tags)
		acc.AddCounter("powercap", map[string]interface{}{
			"energy_uj": zone.energyUj,
		}, tags)
	}

	for _, child := range zone.children {
		g.GatherZone(acc, child)
	}
}

func (g *Powercap) DiscoverZones(prefix string, parent *Zone) (map[string]*Zone, error) {
	zones := make(map[string]*Zone, 0)

	dirs, err := filepath.Glob(path.Join(prefix, fmt.Sprintf("%s:*", g.ControlType)))
	if err != nil {
		return zones, err
	}

	for _, dir := range dirs {
		_, zoneId := filepath.Split(dir)
		zoneId = strings.TrimPrefix(zoneId, fmt.Sprintf("%s:", g.ControlType))

		name, err := readStringFromFile(filepath.Join(dir, "name"))
		if err != nil {
			g.Log.Warn("Failed to read zone for id ", zoneId)
			continue
		}

		zone := Zone{
			path: dir,
			id: zoneId,
			name: strings.TrimSuffix(name, "\n"),
			parent: parent,
			fds: make(map[string]*os.File, 0),
		}

		children, err := g.DiscoverZones(dir, &zone)
		if err != nil {
			g.Log.Warn("Failed to read children for ", zoneId)
		}
		zone.children = children

		var props = [...]string{
			"enabled",
			"energy_uj",
		}

		var failed = false
		for _, prop := range props {
			fd, err := os.Open(filepath.Join(dir, prop))
			if err != nil {
				g.Log.Warn("Failed to load property of zone ", zoneId, ": ", filepath.Join(dir, prop))
				failed = true
				break
			}

			zone.fds[prop] = fd
		}

		if ! failed {
			zones[zoneId] = &zone
		}
	}

	return zones, nil
}

func UpdateZones(zone *Zone) error {
	err := zone.Update()
	if err != nil {
		return err
	}

	for _, child := range zone.children {
		err := UpdateZones(child)
		if err != nil {
			return err
		}
	}

	return nil
}

func (z *Zone) Update() error {
	now := time.Now().UnixNano()

	v, err := readUintFromFile(z.fds["enabled"])
	if err != nil {
		return err
	}
	z.enabled = v != 0

	v, err = readUintFromFile(z.fds["energy_uj"])
	if err != nil {
		return err
	}

	z.powerUw = convertToPower(v - z.energyUj, now - z.tsc)
	z.energyUj = v
	z.tsc = now

	return nil
}

func init() {
	inputs.Add("powercap", func() telegraf.Input { return &Powercap{} })
}

func readStringFromFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func readUintFromFile(fd *os.File) (uint64, error) {
	buffer := make([]byte, 22)

	n, err := fd.ReadAt(buffer, 0)
	if err != nil && err != io.EOF {
		return 0, fmt.Errorf("error on reading file, err: %v", err)
	}

	return strconv.ParseUint(string(buffer[:n-1]), 10, 64)
}

func convertToPower(energy uint64, durationNs int64) uint64 {
	duration := float64(durationNs) / 1_000_000_000
	return uint64(float64(energy) * duration)
}