package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

// NotAvailable human readable
var NotAvailable = "not available"

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
