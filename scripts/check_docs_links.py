#!/usr/bin/env python3
import os
import re
import sys

broken = False
scan_dirs = ["docs"]

for scan_dir in scan_dirs:
    if not os.path.exists(scan_dir):
        continue
    for root, _, files in os.walk(scan_dir):
        for file in files:
            if file.endswith(".md") or file.endswith(".mdx"):
                p = os.path.join(root, file)
                try:
                    content = open(p, encoding="utf-8").read()
                except Exception as e:
                    print(f"Error reading {p}: {e}")
                    continue
                for _, link in re.findall(r"\[([^\]]+)\]\(([^)]+)\)", content):
                    if link.startswith("http") or link.startswith("#") or link.startswith("mailto:"):
                        continue
                    clean_link = link.split("#")[0].split("?")[0]
                    if not clean_link:
                        continue
                    tgt = os.path.normpath(os.path.join(root, clean_link))
                    if not os.path.exists(tgt):
                        print(f"BROKEN LINK: {p} -> {link}")
                        broken = True

if broken:
    sys.exit(1)
print("Documentation link integrity verified.")
