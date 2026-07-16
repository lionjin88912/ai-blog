#!/usr/bin/env python3
"""Build build/ai-blog.ico from build/icon-master.png using Pillow.

Pillow writes classic BMP/DIB icon entries (not PNG) for the small sizes, which
is what Windows Explorer needs to render the icon — a PNG-only .ico shows up as
the blank default icon at 16/32 px. Usage: python3 mkico.py <build-dir>
"""
import os
import sys

from PIL import Image

build = sys.argv[1] if len(sys.argv) > 1 else "build"
master = Image.open(os.path.join(build, "icon-master.png")).convert("RGBA")
out = os.path.join(build, "ai-blog.ico")
master.save(
    out,
    format="ICO",
    sizes=[(16, 16), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)],
    bitmap_format="bmp",
)
print("wrote", out, os.path.getsize(out), "bytes")
