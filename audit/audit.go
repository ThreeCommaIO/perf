package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// NotAvailable human readable
var NotAvailable = "not available"

// Audit holds the key settings of the system
type Audit struct {
	Sysctl    map[string]string `json:"sysctl"`
	Proc      Proc              `json:"proc"`
	Dmesg     string            `json:"dmesg"`
	THP       THP               `json:"transparent_huge_pages"`
	Memory    string            `json:"memory"`
	Disk      Disk              `json:"disk"`
	Network   Network           `json:"network"`
	Distro    Distro            `json:"distro"`
	PowerMgmt PowerMgmt         `json:"power_mgmt"`
}

// Proc holds the key settings of proc interface
type Proc struct {
	Cpuinfo        string            `json:"cpuinfo"`
	Cmdline        string            `json:"cmdline"`
	NetSoftnetStat string            `json:"net/softnet_stat"`
	Cgroups        string            `json:"cgroups"`
	Uptime         string            `json:"uptime"`
	Vmstat         map[string]string `json:"vmstat"`
	Loadavg        string            `json:"loadavg"`
	Zoneinfo       string            `json:"zoneinfo"`
	Partitions     string            `json:"partitions"`
	Version        string            `json:"version"`
}

// THP handles the transparent huge pages
type THP struct {
	Enabled string `json:"enabled"`
	Defrag  string `json:"defrag"`
}

// Disk handles the disk subystem
type Disk struct {
	Scheduler  map[string]string `json:"scheduler"`
	NumDisks   string            `json:"number_of_disks"`
	Partitions string            `json:"partitions"`
}

// Network handles the networking settings
type Network struct {
	Ifconfig string `json:"ifconfig"`
	IP       string `json:"ip"`
	Netstat  string `json:"netstat"`
	SS       string `json:"ss"`
}

// Distro handles the linux distro settings
type Distro struct {
	Issue   string `json:"issue"`
	Release string `json:"release"`
}

// PowerMgmt handles the power management settings
type PowerMgmt struct {
	MaxCState string `json:"max_cstate"`
}

// getSysctl return output from sysctl command
func getSysctl() map[string]string {
	out, err := exec.Command("sysctl", "-a").Output()
	kv := make(map[string]string)
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, pair := range lines {

			if strings.ContainsAny(pair, ":") {
				z := strings.Split(pair, ":")
				kv[z[0]] = strings.TrimSpace(z[1])
			}

			if strings.ContainsAny(pair, "=") {
				z := strings.Split(pair, "=")
				kv[z[0]] = strings.TrimSpace(z[1])
			}
		}
	}
	return kv
}

// read_command captures the output of a command
func read_command(name string, arg ...string) string {
	out, err := exec.Command(name, arg...).Output()
	if err != nil {
		return NotAvailable
	}
	return strings.TrimSpace(string(out))
}

// read_file captures the contents of a file
func read_file(filename string) string {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return NotAvailable
	}
	return strings.TrimSpace(string(dat))
}

// delimited_data splits data by a delimiter
func delimited_data(delimiter string, data string) map[string]string {
	kv := make(map[string]string)

	if strings.Contains(data, NotAvailable) {
		return kv
	}

	lines := strings.Split(data, "\n")
	for _, pair := range lines {
		if strings.ContainsAny(pair, delimiter) {
			z := strings.Split(pair, delimiter)
			key := strings.TrimSpace(z[0])
			value := strings.TrimSpace(z[1])
			kv[key] = value
		}
	}
	return kv
}

// get_release checks with linux distro it is
func get_release() string {
	rels := []string{
		"/etc/SuSE-release", "/etc/redhat-release", "/etc/redhat_version",
		"/etc/fedora-release", "/etc/slackware-release",
		"/etc/slackware-version", "/etc/debian_release", "/etc/debian_version",
		"/etc/os-release", "/etc/mandrake-release", "/etc/yellowdog-release",
		"/etc/sun-release", "/etc/release", "/etc/gentoo-release",
		"/etc/system-release", "/etc/lsb-release",
	}
	for _, path := range rels {
		if _, err := os.Stat(path); err == nil {
			return read_file(path)
		}
	}
	return NotAvailable
}

// get_scheduler handles capturing scheduler data for each block device
func get_scheduler() map[string]string {
	kv := make(map[string]string)
	files, err := ioutil.ReadDir("/sys/block")
	if err != nil {
		return kv
	}
	for _, f := range files {
		block := f.Name()
		path := "/sys/block/" + block + "/queue/scheduler"
		if _, err := os.Stat(path); err == nil {
			kv[block] = read_file(path)
		}
	}
	return kv
}

func main() {
	uid := os.Getuid()

	if uid != 0 {
		log.Fatal("This script must be run as the user root")
	}

	jsonData := Audit{
		Sysctl: getSysctl(),
		Proc: Proc{
			Cpuinfo:        read_file("/proc/cpuinfo"),
			Cmdline:        read_file("/proc/cmdline"),
			NetSoftnetStat: read_file("/proc/net/softnet_stat"),
			Cgroups:        read_file("/proc/cgroups"),
			Uptime:         read_file("/proc/uptime"),
			Vmstat:         delimited_data(" ", read_file("/proc/vmstat")),
			Loadavg:        read_file("/proc/loadavg"),
			Zoneinfo:       read_file("/proc/zoneinfo"),
			Partitions:     read_file("/proc/partitions"),
			Version:        read_file("/proc/version"),
		},
		Dmesg: read_command("dmesg"),
		THP: THP{
			Enabled: read_file("/sys/kernel/mm/transparent_hugepage/enabled"),
			Defrag:  read_file("/sys/kernel/mm/transparent_hugepage/defrag"),
		},
		Memory: read_command("free", "-m"),
		Disk: Disk{
			Scheduler:  get_scheduler(),
			Partitions: read_command("df", "-h"),
			NumDisks:   read_command("lsblk"),
		},
		Network: Network{
			Ifconfig: read_command("ifconfig"),
			IP:       read_command("ip", "addr", "show"),
			Netstat:  read_command("netstat", "-an"),
			SS:       read_command("ss", "-tan"),
		},
		Distro: Distro{
			Issue:   read_file("/etc/issue"),
			Release: get_release(),
		},
		PowerMgmt: PowerMgmt{
			MaxCState: read_file("/sys/module/intel_idle/parameters/max_cstate"),
		},
	}
	b, _ := json.MarshalIndent(jsonData, "", "  ")
	hostname, _ := os.Hostname()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	auditFile := fmt.Sprintf("audit-%s-%s.json", hostname, timestamp)

	err := ioutil.WriteFile(auditFile, b, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Generated audit file: %s\n", auditFile)
}
