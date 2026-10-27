package main

import (
	"fmt"
)

// More tools for the chat panel: effects and color adjustments, scaling and turning, laying in a picture from a file, the drawing extras, what changed
// against the file, the folder and the fonts, and going back in the history. They are told to the model in mcpTools (see moreTools) and run from callTool.

var (
	pEffects = []string{"blur", "pixelize", "emboss", "explosion", "grayscale", "negative", "sepia", "censor", "fill", "gradient",
		"brightness", "contrast", "saturation", "hue", "temperature", "gamma", "levels", "auto_levels", "sharpen"}
	pFeather = prop("integer", "Pixels over which the edge of the part fades into the rest (0 = hard edge)")
	pShadow  = map[string]any{"type": "object", "description": "A soft shadow under the shape: {dx, dy, blur, color} (default 3, 3, 4, black)", "properties": map[string]any{
		"dx": map[string]any{"type": "integer"}, "dy": map[string]any{"type": "integer"}, "blur": map[string]any{"type": "integer"}, "color": map[string]any{"type": "string"}}}
	pOpacity = prop("number", "0 (invisible) to 1 (solid, the default)")
)

func moreTools() []toolDef {
	return []toolDef{
		{"apply_effect", "Applies an effect or a color adjustment to part of the picture: a box (x, y, w, h), a polygon (any number of points), a mask, what the user selected (selection: true), or, when none of these is given, all of the picture. Effects: blur (radius), pixelize (size = the side of a block), emboss (angle = where the light comes from, amount), explosion (a burst from cx, cy, amount 0 to 100), grayscale, negative, sepia, censor (a mosaic too coarse to read, or bar: true for a solid bar; use it to hide faces, plates, names, screens), fill (color, opacity), gradient (color to color2, angle, kind linear or radial, opacity), sharpen (amount 0 to 500, radius), and the adjustments brightness, contrast, saturation, temperature (amount -100 to 100; temperature above 0 is warmer), hue (angle in degrees), gamma (gamma, 1 = no change), levels (black, white, gamma), auto_levels. One step of the history, so undo takes it back; use feather for a soft edge. For a polygon of the outline of a thing use create_mask or select_by_color first and pass the mask.", schema([]string{"effect"}, map[string]any{
			"effect": map[string]any{"type": "string", "enum": pEffects},
			"x":      pX, "y": pY, "w": prop("integer", "width of the box in pixels"), "h": prop("integer", "height of the box in pixels"), "polygon": pPoly, "mask": pMask,
			"selection": prop("boolean", "Work on what the user selected in the window (the polygon, else the box)"), "feather": pFeather,
			"radius": prop("number", "blur and sharpen: pixels (blur default 6, sharpen 2)"), "size": prop("integer", "pixelize and censor: the side of a block in pixels (default 12; censor: about a sixth of the part)"),
			"amount": prop("number", "The strength: -100 to 100 for the adjustments, 0 to 100 for emboss and explosion, 0 to 500 for sharpen"),
			"angle":  prop("number", "degrees: emboss (the light), hue (the turn), gradient (0 = left to right, 90 = top to bottom)"),
			"cx":     prop("number", "explosion: x of the middle of the burst (default the middle of the part)"), "cy": prop("number", "explosion: y of the middle"),
			"color": prop("string", "fill, gradient (start) and the bar of censor: a color (#rrggbb, #rrggbbaa, a name)"), "color2": prop("string", "gradient: the end color"),
			"kind": map[string]any{"type": "string", "enum": []string{"linear", "radial"}, "description": "gradient"}, "opacity": pOpacity,
			"bar": prop("boolean", "censor: a solid bar (color, black by default) instead of a mosaic"), "gamma": prop("number", "gamma and levels: above 1 lightens the middle tones"),
			"black": prop("integer", "levels: the value (0 to 254) that becomes black"), "white": prop("integer", "levels: the value (1 to 255) that becomes white"), "snapshot": pSnap})},
		{"resize_image", "Scales the picture to a new size (resampling the pixels, unlike resize_canvas which adds or cuts room). Give w and h, or only one of them to keep the proportions, or scale (2 = twice as big). One step of the history.", schema(nil, map[string]any{
			"w": prop("integer", "new width in pixels"), "h": prop("integer", "new height in pixels"), "scale": prop("number", "factor instead of w and h (0.5 = half)"),
			"filter": map[string]any{"type": "string", "enum": []string{"smooth", "sharp", "nearest"}, "description": "smooth (default, best for photos), sharp, nearest (whole pixels, for pixel art)"}, "snapshot": pSnap})},
		{"rotate_image", "Turns the picture clockwise by angle degrees. Quarter turns (90, 180, 270) lose nothing. Other angles make the canvas bigger to hold the corners (the corners are transparent where the file type can be, else white, or bg), or with crop keep the size and cut the corners off. One step of the history.", schema([]string{"angle"}, map[string]any{
			"angle": prop("number", "degrees clockwise (negative: counter-clockwise)"), "bg": prop("string", "Color of the new corners"), "crop": prop("boolean", "Keep the size of the picture (default false)"), "snapshot": pSnap})},
		{"flip_image", "Mirrors the picture: horizontal (left to right) and/or vertical (top to bottom). One step of the history.", schema(nil, map[string]any{
			"horizontal": prop("boolean", "Mirror left to right"), "vertical": prop("boolean", "Mirror top to bottom"), "snapshot": pSnap})},
		{"place_image", "Lays a picture from a file (png, jpg, gif, bmp, tif, webp) over this picture, its top left corner at (x, y), optionally scaled and see-through (a logo, a watermark, a sticker). What hangs over the edge is cut off. One step of the history.", schema([]string{"path", "x", "y"}, map[string]any{
			"path": prop("string", "Full path of the picture file"), "x": pX, "y": pY, "scale": prop("number", "1 = the size it has"), "w": prop("integer", "or the width to scale it to (the height follows)"), "h": prop("integer", "or the height (the width follows)"),
			"opacity": pOpacity, "snapshot": pSnap})},
		{"draw_polygon", "Draws a polygon through the points (closed): an outline, or filled with fill_color (also with the outline). Good for arrows, stars, tails, any shape. Can be see-through (opacity) and have a shadow.", schema([]string{"polygon"}, map[string]any{
			"polygon": pPoly, "color": pColor, "width": pWidth, "fill_color": prop("string", "Fill the inside with this color"), "style": map[string]any{"type": "string", "enum": []string{"solid", "dashed", "dotted"}},
			"opacity": pOpacity, "shadow": pShadow, "snapshot": pSnap})},
		{"draw_callout", "Draws a speech bubble: a rounded box at (x, y), w by h, with a tail that points at (tail_x, tail_y), filled (white by default) and outlined. The answer tells where to put the text (add_text with box_w and align center).", schema([]string{"x", "y", "w", "h", "tail_x", "tail_y"}, map[string]any{
			"x": pX, "y": pY, "w": prop("integer", "width of the bubble"), "h": prop("integer", "height of the bubble"), "tail_x": prop("integer", "x the tail points at"), "tail_y": prop("integer", "y the tail points at"),
			"radius": prop("integer", "corner radius (default 12)"), "color": pColor, "fill_color": prop("string", "Fill color (default white)"), "width": pWidth, "opacity": pOpacity, "shadow": pShadow, "snapshot": pSnap})},
		{"clear_selection", "Clears (cuts out) a part of the picture: fills it with the window's second (background) color, the way Delete does in Paint. By default it is the part the user selected in the window (the polygon, else the box); or give a box (x, y, w, h), a polygon or a mask. color overrides the second color. One step of the history.", schema(nil, map[string]any{
			"x": pX, "y": pY, "w": prop("integer", "width of the box in pixels"), "h": prop("integer", "height of the box in pixels"), "polygon": pPoly, "mask": pMask,
			"color": prop("string", "Fill with this color instead of the window's second color (a name, #rrggbb or #rrggbbaa; transparent makes a hole)"), "feather": pFeather, "snapshot": pSnap})},
		{"diff_original", "Compares the picture as it is now with the file on the disk: whether it changed, how many pixels, the box of the change and how strong it is on average. Use it to check that an edit touched only what you meant. Does not change the picture.", schema(nil, map[string]any{})},
		{"list_folder", "Lists the pictures in the folder of the open file, by file name (not the path), and where the current one is among them. Does not change anything.", schema(nil, map[string]any{})},
		{"list_fonts", "Lists font families that add_text can be asked for (the common ones, and the names of the font files found on this computer). Does not change anything.", schema(nil, map[string]any{})},
		{"goto_step", "Goes to a step of the undo history (the numbers come from history): undoes or redoes as many steps as it takes. Step 0 is the start. One call instead of many undo calls.", schema([]string{"step"}, map[string]any{"step": prop("integer", "the step number"), "snapshot": pSnap})},
		{"checkpoint", "Remembers the picture as it is now under a name, so that you can try something and come back with goto_checkpoint. The user's and your later changes are in the history; a checkpoint is lost when the step it points at is forgotten (a new change after an undo drops the steps that were undone).", schema([]string{"name"}, map[string]any{"name": prop("string", "a short name")})},
		{"goto_checkpoint", "Goes back to (or forward to) a named checkpoint made with checkpoint.", schema([]string{"name"}, map[string]any{"name": prop("string", "the name"), "snapshot": pSnap})},
		{"show_area", "Scrolls and zooms the window so that the part x, y, w, h fills it, so that the user sees what you are talking about. With mark: true it is also selected (marching dashes). Does not change the picture.", schema([]string{"x", "y", "w", "h"}, map[string]any{
			"x": pX, "y": pY, "w": prop("integer", "width of the part"), "h": prop("integer", "height of the part"), "mark": prop("boolean", "Also select it (default false)")})},
	}
}

// moreCall runs a tool of moreTools; handled says whether the name was one of them.
func (h *chatHub) moreCall(name string, args map[string]any) (content []any, handled bool, err error) {
	switch name {
	case "apply_effect":
		content, err = h.effectTool(args)
	case "clear_selection":
		content, err = h.clearTool(args)
	case "resize_image", "rotate_image", "flip_image", "place_image":
		content, err = h.transformTool(name, args)
	case "diff_original", "list_folder", "list_fonts":
		cmd := map[string]string{"diff_original": "diff", "list_folder": "folder", "list_fonts": "fonts"}[name]
		var res map[string]any
		if res, err = h.window(map[string]any{"cmd": cmd}); err == nil {
			delete(res, "ok")
			content = []any{jsonText(res)}
		}
		return content, true, err
	case "goto_step":
		var n int
		if n, err = intArg(args, "step"); err == nil {
			_, err = h.window(map[string]any{"cmd": "goto", "step": n})
			content = []any{textContent(fmt.Sprintf("Went to step %d of the history", n))}
		}
	case "checkpoint":
		if str(args, "name") == "" {
			return nil, true, fmt.Errorf("\"name\" is missing")
		}
		var res map[string]any
		if res, err = h.window(map[string]any{"cmd": "checkpoint", "name": str(args, "name")}); err == nil {
			content = []any{textContent(fmt.Sprintf("Checkpoint %q is step %v of the history", str(args, "name"), res["step"]))}
		}
		return content, true, err
	case "goto_checkpoint":
		if _, err = h.window(map[string]any{"cmd": "goto", "checkpoint": str(args, "name")}); err == nil {
			content = []any{textContent(fmt.Sprintf("Went back to the checkpoint %q", str(args, "name")))}
		}
	case "show_area":
		var v []int
		if v, err = ints(args, "x", "y", "w", "h"); err == nil {
			if v[2] < 1 || v[3] < 1 {
				return nil, true, fmt.Errorf("w and h must be 1 or more")
			}
			_, err = h.window(map[string]any{"cmd": "show", "x": v[0], "y": v[1], "w": v[2], "h": v[3], "mark": args["mark"] == true})
			content = []any{textContent("The window shows that part now")}
		}
		return content, true, err
	default:
		return nil, false, nil
	}
	if err != nil {
		return nil, true, fmt.Errorf("%s: %w", name, err)
	}
	if args["snapshot"] == true {
		if snap, serr := h.snapshotTool(map[string]any{}); serr == nil {
			content = append(content, snap...)
		}
	}
	return content, true, nil
}

func (h *chatHub) effectTool(args map[string]any) ([]any, error) {
	op := str(args, "effect")
	if op == "" {
		return nil, fmt.Errorf("\"effect\" is missing")
	}
	cmd := map[string]any{"cmd": "effect", "op": op}
	for _, k := range []string{"x", "y", "w", "h", "feather", "radius", "size", "amount", "angle", "cx", "cy", "kind", "opacity", "bar", "gamma", "black", "white", "mask", "selection"} {
		if v, ok := args[k]; ok {
			cmd[k] = v
		}
	}
	for _, k := range []string{"color", "color2"} {
		if s := str(args, k); s != "" {
			c, err := parseColorAlpha(s, "")
			if err != nil {
				return nil, err
			}
			cmd[k] = c
		}
	}
	pts, err := polygonArg(args)
	if err != nil {
		return nil, err
	}
	if pts != nil {
		cmd["pts"] = pts
	}
	res, err := h.window(cmd)
	if err != nil {
		return nil, fmt.Errorf("apply_effect: %w", err)
	}
	note, _ := res["note"].(string)
	if note == "" {
		note = "Applied " + op
	}
	return []any{textContent(note)}, nil
}

func (h *chatHub) clearTool(args map[string]any) ([]any, error) {
	cmd := map[string]any{"cmd": "effect", "op": "fill", "clear": true}
	for _, k := range []string{"x", "y", "w", "h", "mask", "feather"} {
		if v, ok := args[k]; ok {
			cmd[k] = v
		}
	}
	pts, err := polygonArg(args)
	if err != nil {
		return nil, err
	}
	if pts != nil {
		cmd["pts"] = pts
	}
	if _, ok := args["w"]; !ok && pts == nil && args["mask"] == nil {
		cmd["selection"] = true // no part given: what the user selected
	}
	if s := str(args, "color"); s != "" {
		c, err := parseColorAlpha(s, "")
		if err != nil {
			return nil, err
		}
		cmd["color"] = c
	}
	if _, err := h.window(cmd); err != nil {
		return nil, err
	}
	return []any{textContent("Cleared the part (filled with the second color)")}, nil
}

func (h *chatHub) transformTool(name string, args map[string]any) ([]any, error) {
	op := map[string]string{"resize_image": "resize", "rotate_image": "rotate", "flip_image": "flip", "place_image": "place"}[name]
	cmd := map[string]any{"cmd": "transform", "op": op}
	for _, k := range []string{"w", "h", "scale", "filter", "angle", "crop", "horizontal", "vertical", "path", "x", "y", "opacity"} {
		if v, ok := args[k]; ok {
			cmd[k] = v
		}
	}
	if s := str(args, "bg"); s != "" {
		c, err := parseColorAlpha(s, "")
		if err != nil {
			return nil, err
		}
		cmd["bg"] = c
	}
	if _, err := h.window(cmd); err != nil {
		return nil, err
	}
	st, err := h.state()
	if err != nil {
		return []any{textContent("Done: " + name)}, nil
	}
	if pw, ph, err := pictureSize(st); err == nil {
		return []any{textContent(fmt.Sprintf("Done: %s. The picture is %d × %d pixels now", name, pw, ph))}, nil
	}
	return []any{textContent("Done: " + name)}, nil
}

// addShapeExtras gives the shapes and the text the options that the extra tools have (corners, see-through, shadow, outline, wrapping).
func addShapeExtras(cmds []map[string]any, args map[string]any) {
	for _, c := range cmds {
		switch c["cmd"] {
		case "shape":
			for _, k := range []string{"radius", "opacity", "shadow"} {
				if v, ok := args[k]; ok {
					c[k] = v
				}
			}
		case "text":
			for _, k := range []string{"align", "box_w", "outline_width", "shadow"} {
				if v, ok := args[k]; ok {
					c[k] = v
				}
			}
			if s := str(args, "outline"); s != "" {
				if oc, err := parseColor(s, ""); err == nil {
					c["outline"] = oc
				}
			}
		}
	}
}
