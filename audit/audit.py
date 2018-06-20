import json
import subprocess
import re


def empty(value):
    try:
        value = float(value)
    except ValueError:
        pass
    return bool(value)


def read_file(filename):
    try:
        return open(filename).read().strip()
    except:
        "not available"


def read_command(command):
    try:
        res = subprocess.check_output(command.split()).strip()
        return res
    except OSError:
        return "not available"
    except:
        return "not available"


def get_sysctl():
    i = {}
    for item in filter(None, read_command('sysctl -a').split("\n")):
        if " = " in item:
            j = item.split(" = ")
        elif ": " in item:
            j = item.split(": ")

        i[j[0].strip()] = j[1].strip()

    return i


def get_release():
    rels = ["/etc/SuSE-release", "/etc/redhat-release", "/etc/redhat_version", "/etc/fedora-release",
        "/etc/slackware-release", "/etc/slackware-version", "/etc/debian_release", "/etc/debian_version",
        "/etc/os-release", "/etc/mandrake-release", "/etc/yellowdog-release", "/etc/sun-release",
        "/etc/release", "/etc/gentoo-release", "/etc/system-release", "/etc/lsb-release"]
    for rel in rels:
        try:
            res = read_file(rel)
        except:
            pass

    return res


def get_scheduler():
    try:
        p1 = subprocess.Popen(['ls', '-l', '/sys/block'], stdout=subprocess.PIPE)
        p2 = subprocess.Popen(['awk', "{print $9}"], stdin=p1.stdout, stdout=subprocess.PIPE)
        blocks = filter(None, p2.communicate()[0].split("\n"))

        res = {}
        for block in blocks:
          res[block] = read_file('/sys/block/' + block + '/queue/scheduler')

        return res
    except:
        return "not available"


def tabular_data(data):
    res = []
    lines = data.split("\n")

    # keys
    first_line = re.sub('Local Address:Port','Local-Address-Port', lines[0])
    first_line = re.sub('Peer Address:Port','Peer-Address-Port', first_line)
    keys = ' '.join(first_line.split()).split(' ')

    try:
        keys.remove('on')
    except:
        pass

    for i, s in enumerate(keys):
        keys[i] = re.sub(r'\W+', '', s.lower().strip())

    # data
    for i in range(1,len(lines)):
        a = {}
        vals = ' '.join(lines[i].split()).split(' ')

        # /proc/partitions has no data in the second row.  If no val present, skip the row
        if vals[0] == "":
          continue

        for j in range(len(keys)):
            key = keys[j]
            try:
                val = vals[j]
            except:
                val = ''
            a[key] = val
        res.append(a)

    return res


def delimited_data(delimiter, data):
    res = {}
    lines = filter(None, data.split("\n"))
    for i in range(len(lines)):
        j = lines[i].split(delimiter)
        res[re.sub(r'\W+', '_', j[0].lower().strip())] = j[1].strip()

    return res


def main():

    a = {}

    # SYSCTL
    a['sysctl'] = get_sysctl()

    # PROC
    a['proc'] = {}
    a['proc']['cpuinfo'] = delimited_data(':', read_file('/proc/cpuinfo'))
    a['proc']['cmdline'] = read_file('/proc/cmdline')
    a['proc']['net/softnet_stat'] = read_file('/proc/net/softnet_stat')
    a['proc']['cgroups'] = tabular_data(read_file('/proc/cgroups'))
    a['proc']['uptime'] = read_file('/proc/uptime')
    a['proc']['vmstat'] = delimited_data(' ', read_file('/proc/vmstat'))
    a['proc']['loadavg'] = read_file('/proc/loadavg')
    a['proc']['zoneinfo'] = read_file('/proc/zoneinfo')
    a['proc']['partitions'] = tabular_data(read_file('/proc/partitions'))
    a['proc']['version'] = read_file('/proc/version')

    # PARTITIONS
    a['disk_partitions'] = tabular_data(read_command('df -h'))

    # DMESG
    a['dmesg'] = read_command('dmesg')

    # THP
    a['transparent_huge_pages'] = {}
    a['transparent_huge_pages']['enabled'] = read_file('/sys/kernel/mm/transparent_hugepage/enabled')
    a['transparent_huge_pages']['defrag'] = read_file('/sys/kernel/mm/transparent_hugepage/defrag')

    # MEMORY
    a['memory'] = read_command('free -m')

    # DISK
    a['disk'] = {}
    a['disk']['scheduler'] = get_scheduler()
    a['disk']['number_of_disks'] = tabular_data(read_command('lsblk'))

    # NETWORK
    a['network'] = {}
    a['network']['ifconfig'] = read_command('ifconfig')
    a['network']['ip'] = read_command('ip addr show')
    a['network']['netstat'] = read_command('netstat -a')
    a['network']['ss'] = tabular_data(read_command('ss -tan'))

    # DISTRO
    a['distro'] = {}
    a['distro']['issue'] = read_file('/etc/issue')
    a['distro']['release'] = get_release()

    # POWER MGMT
    a['power_mgmt'] = {}
    a['power_mgmt']['max_cstate'] = read_file('/sys/module/intel_idle/parameters/max_cstate')

    print json.dumps(a, indent=4)


if __name__ == "__main__":
   main()
