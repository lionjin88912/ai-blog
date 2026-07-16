#!/usr/bin/env python3
"""Pack build/i-<size>.png files into build/ai-blog.ico (multi-resolution,
PNG-embedded). Usage: python3 mkico.py <build-dir>"""
import os
import struct
import sys

build = sys.argv[1] if len(sys.argv) > 1 else "build"
sizes = [16, 32, 48, 64, 128, 256]
imgs = []
for s in sizes:
    with open(os.path.join(build, f"i-{s}.png"), "rb") as f:
        imgs.append((s, f.read()))

out = os.path.join(build, "ai-blog.ico")
with open(out, "wb") as f:
    f.write(struct.pack("<HHH", 0, 1, len(imgs)))  # ICONDIR
    off = 6 + 16 * len(imgs)
    for s, data in imgs:
        w = 0 if s >= 256 else s  # 0 means 256 in the ICO spec
        f.write(struct.pack("<BBBBHHII", w, w, 0, 0, 1, 32, len(data), off))
        off += len(data)
    for _, data in imgs:
        f.write(data)
print("wrote", out, os.path.getsize(out), "bytes")
