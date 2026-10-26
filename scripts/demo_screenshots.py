#!/usr/bin/env python3
"""Makes the screenshots of the manual with mcView in control mode (macOS; docs/control.md).

    scripts/demo_screenshots.py SAMPLES_DIR [OUT_DIR] [APP]

SAMPLES_DIR holds berries.jpg, coast.jpg and lake.jpg (free pictures, no people in them); OUT_DIR defaults to docs/screenshots (WebP when Pillow is installed, else PNG), APP to
build/bin/mcview.app. The window is a fixed size, so the pictures look alike from one run to the next.
"""
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from mcview_ctl import McView  # noqa: E402


def main(samples, out="docs/screenshots", app="build/bin/mcview.app"):
    mc = McView.start(app, os.path.join(samples, "berries.jpg"))

    def shot(name):
        png = os.path.join(out, name + ".png")
        mc.shot(png)
        try:  # WebP keeps the transparent corners and the shadow, at a tenth of the size (Pillow); without it the PNG stays
            from PIL import Image
            Image.open(png).save(png[:-4] + ".webp", "WEBP", quality=90, method=6)
            os.remove(png)
        except ImportError:
            pass
        print("  ", name)

    def px(fx, fy):  # a place of the picture, as fractions of its size
        w, h = mc.state()["size"]
        return [round(fx * w), round(fy * h)]

    def open_picture(name, mode=2):
        mc.cmd("open", path=os.path.join(samples, name))
        mc.cmd("mode", mode=mode)

    mc.cmd("window", w=1500, h=960)
    mc.cmd("mode", mode=2)
    shot("01-view")

    open_picture("lake.jpg", 5)  # the side edges touch the window
    shot("02-zoom-fit-width")
    mc.cmd("mode", mode=6)  # the top and bottom edges touch it
    shot("03-zoom-fit-height")

    open_picture("coast.jpg", 2)
    mc.cmd("edit")
    shot("04-edit-mode")

    # draw with the tools, in the coordinates of the picture
    mc.cmd("color", fg="#ED1C24")
    mc.cmd("tool", name="pen")
    mc.cmd("option", width=14)
    ring = [px(0.78, 0.36), px(0.79, 0.31), px(0.825, 0.295), px(0.862, 0.315), px(0.866, 0.37), px(0.83, 0.405), px(0.79, 0.385), px(0.78, 0.35)]  # round the palm trees on the cliff
    mc.cmd("shape", kind="path", pts=ring, color="#ED1C24", width=14)
    mc.cmd("color", fg="#FFF200")
    mc.cmd("tool", name="ellipse")
    mc.cmd("option", width=10, style="solid", fill="outline")
    a, b = px(0.60, 0.53), px(0.72, 0.66)
    mc.cmd("shape", kind="ellipse", x0=a[0], y0=a[1], x1=b[0], y1=b[1], color="#FFF200", width=10, style="solid", fill="outline")
    mc.cmd("color", fg="#FFFFFF")
    mc.cmd("tool", name="rect")
    mc.cmd("option", width=8, style="dashed", fill="outline")
    a, b = px(0.30, 0.10), px(0.58, 0.30)
    mc.cmd("shape", kind="rect", x0=a[0], y0=a[1], x1=b[0], y1=b[1], color="#FFFFFF", width=8, style="dashed", fill="outline")
    mc.cmd("tool", name="line")
    mc.cmd("option", width=10, style="solid")
    a, b = px(0.24, 0.50), px(0.40, 0.44)
    mc.cmd("shape", kind="line", x0=a[0], y0=a[1], x1=b[0], y1=b[1], color="#FFFFFF", width=10, style="solid")
    mc.cmd("color", fg="#FFF200")
    mc.cmd("option", font="system-ui, sans-serif", size=190, bold=True)
    mc.cmd("text", x=px(0.06, 0)[0], y=px(0, 0.80)[1], value="Evening walk", color="#FFF200")
    mc.cmd("tool", name="rect")  # the ribbon of a shape tool stays on the screen
    shot("05-edit-draw")

    # zoom: look at the details, and draw exactly
    people = [2548, 1425]  # the walkers on the beach (coast.jpg is 3840 × 2563)
    mc.cmd("tool", name="select")
    mc.cmd("zoom", scale=8, x=people[0], y=people[1])
    mc.cmd("pixel", x=people[0] + 6, y=people[1] + 3)
    shot("06-zoom-pixels")
    mc.cmd("tool", name="pen")
    mc.cmd("color", fg="#ED1C24")
    mc.cmd("option", width=2)
    mc.cmd("zoom", scale=4, x=people[0], y=people[1])
    mc.cmd("shape", kind="path", pts=[[people[0] - 70, people[1] + 70], [people[0] - 20, people[1] + 50], [people[0] + 40, people[1] + 70]], color="#ED1C24", width=2)
    shot("07-zoom-draw")

    # the colors
    mc.cmd("mode", mode=2)
    mc.cmd("palette", open="fg")
    shot("08-palette")
    mc.cmd("palette")

    # a selection, which is on the clipboard at once
    mc.cmd("tool", name="select")
    a, b = px(0.52, 0.30), px(0.80, 0.62)
    mc.cmd("select", x0=a[0], y0=a[1], x1=b[0], y1=b[1])
    shot("09-selection")

    # the home screen
    mc.answer("discard", "discard")
    mc.cmd("menu", name="close-image")
    shot("10-home")
    mc.quit()


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(__doc__)
        sys.exit(2)
    main(*sys.argv[1:4])
