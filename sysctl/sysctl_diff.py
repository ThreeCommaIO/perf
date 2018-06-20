#!/usr/bin/env python
"""
Compare outputs of sysctl from different machines

$ ./sysctl_diff.py <f1> <f2>

"""
import sys


def parse_sysctl(filename):
    kv = {}
    with open(filename) as f:
        for line in f.readlines():
            try:
                k, v = line.strip().split('=')
                kv[k.strip()] = v.strip()
            except:
                pass
    return kv


def main():
    f1 = sys.argv[1]
    f2 = sys.argv[2]

    f1map = parse_sysctl(f1)
    f2map = parse_sysctl(f2)

    keys = list(sorted(f1map.keys()) + sorted(f2map.keys()))
    for key in keys:
        if key in f1map and key in f2map:
            if f1map[key] != f2map[key]:
                print "%s (%s => %s)" % (key, f1map[key], f2map[key])


if __name__ == '__main__':
    main()
