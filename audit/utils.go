package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

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
