# Chat panel

The speech bubble at the bottom right of a picture window (or the key **C**) opens a chat with Claude about the picture of that window. Ask it to
look at the picture, to draw on it, to change it; the changes appear in the window and are steps of its undo history.

## How it works

mcView starts the **Claude Code** that is installed on the computer (`claude -p`, with streamed JSON output) in the background and gives it the
picture tools of this window, and the web:

- no terminal has to be open, and no API key is used: Claude Code works with the user's **Claude subscription**;
- Claude can use the mcView tools (below) and web search and fetch, and nothing else: no files, no shell (`--tools "WebSearch,WebFetch"`, only `mcp__mcview` and those two allowed). **AI → Web search and fetch** turns the web off (it is on by default, and the choice is kept in `ai.json` in the settings folder). A page it reads is untrusted text, so it can be asked to look things up, but it is not given anything that reaches outside the picture except the web;
- the tools come from a **private MCP server** of the window: it listens on `127.0.0.1` on a free port, only while the panel has been used, and only
  answers requests with a token that exists in this program and, for the length of one answer, in a temporary file (mode 0600) that Claude Code
  reads. It refuses requests with an `Origin` header (a web page) or another host. There is no setting for it and no other program is given access;
- the conversation goes on between messages (`--resume`); **New** starts another one; **Stop** ends the answer that is being written;
- Claude Code runs in an empty folder of the temp directory.

## Signing in

The panel asks `claude auth status`. When the user is not signed in it offers **Sign in with Claude**, which runs `claude auth login --claudeai`:
Claude Code opens a browser tab, and keeps the token it gets there in its own storage. mcView never sees the password or the token. If the
browser does not open, the address is shown in the panel. When the sign-in is over the panel checks again by itself.

The panel shows the account and the plan in its header. A sign-in that is not a Claude subscription (an API key or Console login) is marked
in red, because that is billed by use.

## What is needed

Claude Code (<https://claude.com/product/claude-code>). The panel says so when it is missing.

## Tools

All positions are pixels of the picture as it is shown (0,0 is the top left pixel). Every drawing is one step of the window's undo history,
and nothing is written to a file until `save_image`. The tools change the picture at once; nothing moves a pointer or presses a key.

| tool | does |
|---|---|
| `window_info` | the picture's name and size, unsaved changes, tool, selection, zoom |
| `image_info` | the file name (not the folder), type and size, the EXIF data of a photo (camera, time, exposure, ISO; only *whether* there is a GPS place), and how the colors are used: the number of colors, the main colors with their share, average color, brightness, transparency |
| `apply_effect` | an effect on a box, a polygon, a mask or the user's selection (or all of the picture): blur, pixelize, emboss, explosion, grayscale, negative, sepia, censor (mosaic or bar), fill, gradient, sharpen, and the adjustments brightness, contrast, saturation, temperature, hue, gamma, levels, auto levels; with a soft edge (`feather`) if wanted |
| `resize_image`, `rotate_image`, `flip_image` | scale the picture (smooth, sharp or whole pixels), turn it by any angle, mirror it |
| `place_image` | lay a picture from a file over this one, scaled and see-through if wanted |
| `draw_polygon`, `draw_callout` | a closed polygon (outline and/or fill), a speech bubble with a tail; `draw_rect` has rounded corners, and the shapes can be see-through and have a shadow |
| `clear_selection` | fill the user's selection (or a box, polygon, mask) with the window's second color, like Delete in Paint |
| `cut_selection`, `crop_to_selection`, `save_selection`, `print_selection` | the right-click menu of a selection, for Claude: cut it, crop the picture to it, save only it to a file, print it |
| `diff_original` | what changed against the file: pixels, box, strength |
| `list_folder`, `list_fonts` | the picture names in the folder (no path), the fonts for `add_text` (which also has align, wrapping, outline, shadow) |
| `goto_step`, `checkpoint`, `goto_checkpoint` | jump in the history, and name a state to come back to |
| `show_area` | scroll and zoom the window to a part (and mark it), so the user sees what is meant |
| `history` | the steps of the undo history, each with what was done and who did it (the user or Claude): the ones undo takes back, and the ones redo brings back |
| `get_image` | the whole picture, with the changes so far, pixel for pixel (a PNG at its real size). A picture of up to 64 × 64 pixels also comes back as text, a palette and one string per row, in the form `set_pixels` takes |
| `snapshot` | for a big picture: a scaled-down view (`max`: longest side, 1568 by default), or a part of it in full detail (`x`,`y`,`w`,`h`); `zoom` (2 to 32) enlarges a small part in whole pixels, `outline` (`[[x, y], …]`) draws a shape over the answer only, `mask` tints a mask, `compare` shows the original file (left) beside the picture now (right) |
| `new_image` | a new blank picture of `width` × `height`, white, a color or `transparent` (refused over unsaved changes unless `discard`) |
| `set_pixels` | single pixels, for pixel art: `pixels` as `{x, y, color}`, or a `grid` of strings with a `palette` (`"R"` → `#ff0000`, `"."` → `transparent`); one undo step, instant |
| `draw_arrow`, `draw_line` | from `x1`,`y1` to `x2`,`y2` (the arrow's head is at the second point) |
| `draw_rect`, `draw_ellipse` | the box with its top left corner at `x`,`y`, `w` × `h`; `fill_color` fills it |
| `draw_path` | a pencil line through `points` |
| `add_text` | `text` with its top left corner at `x`,`y`; `size`, `bold`, `italic`, `font` |
| `resize_canvas` | `width` × `height`: a bigger (or smaller) canvas without scaling the picture. `anchor` (default `bottom-left`: the new room is above and to the right; `top-left`, `center`… also work) says where the picture stays, `background` colors the new room (transparent, or white for JPEG and BMP). One undo step |
| `crop` | keeps only the part `x`,`y`,`w`,`h` of the picture. One undo step |
| `select_by_color` | a **mask** (an id such as `m1`) of the pixels with the color of the pixel `x`,`y`, give or take `tolerance`: the connected area (default) or all of them (`contiguous: false`); holes inside are filled; `grow`, `shrink`. Does not change the picture; a `snapshot` with the id tints the area |
| `create_mask`, `edit_mask` | a mask of a `polygon` (as many points as the outline needs), several `polygons` joined into one mask, or a box; `edit_mask` adds, subtracts or intersects masks, inverts, fills holes, grows or shrinks. A mask is for the size of the picture it was made for |
| `selection_mask` | shows the model what the user selected in the window (the polygon of the polygon select tool, or else the rectangle of the select tool): the corners, a mask id for the other tools, and two pictures: the whole picture with the outline, and the selected part enlarged and tinted, with its surroundings |
| `select_polygon` | selects a polygon in the window so that the user sees it (and can copy or delete it); an empty polygon lets it go |
| `discard_changes` | the picture goes back to the file on the disk (or the pasted or new picture); one undo step |
| `clone_region` | copies a part (a `mask`, a `polygon`, or the box `x`,`y`,`w`,`h`) so that its top left corner lands on `to_x`,`to_y`; `flip_h`, `flip_v`, `rotate`, any `angle` and `scale` (smooth, around its middle), `feather` (soft edge, pixels), `opacity`, `blend: "seamless"` (healing brush: the part keeps its texture and takes the light and color of the place). The answer warns when a part is copied again from the same place. One undo step |
| `seam_score` | how much the edge of an area shows: a seam ratio (about 1 cannot be seen, over 2 shows, over 3 is obvious), the step in brightness, the worst sides. Does not change the picture; `blend_edges` and `clone_region` also tell it |
| `blend_edges` | hides the seam around a patch (box or `polygon`): `method` `smooth` (blur the edge band), `color_match` (move the colors inside to the average and spread around) or `poisson` (spread the difference at the edge smoothly over the inside); `width`, `strength` |
| `fill_region` | fills an area (a `mask`, a `polygon` or a box) from the pixels around it. `method`: `patch` (the default: patches copied from the surroundings, edge inward, `patch` size; continues textures), `smooth` (the colors of the edge spread inward smoothly: paper, post-its, walls, skin, sky), `nearest` (each pixel takes the nearest color outside), `average` (one flat color around), `solid` (`color`); `grain` adds the noise of the surroundings so that a smooth fill does not look flat. `source_mask` limits what may be used, `avoid_mask` what may not. Does not invent objects |
| `flood_fill` | the paint bucket, with a `tolerance` |
| `select_region`, `copy`, `paste` | select a part (it is put on the clipboard), copy, paste |
| `undo`, `redo` | |
| `get_pixel` | the color at `x`,`y` |
| `open_image`, `save_image` | open a file (refused over unsaved changes unless `discard`); save to a full path (refused over an existing file unless `overwrite`) |

A small new picture is shown big, in whole pixels, so that you can watch the sprite appear. The drawing tools take `color` (`#rrggbb`, `#rgb` or a name, red by default) and `width` (by default about 1/250 of the picture's longer side);
`snapshot: true` returns a snapshot of the result with the answer.

A small new picture is shown big, in whole pixels, so that you can watch the sprite appear. The drawing tools take `color` (`#rrggbb`, `#rgb` or a name, red by default) and `width` (by default about 1/250 of the picture's longer side);
`snapshot: true` returns a snapshot of the result with the answer.

## Limits of the prototype

- Claude Code still reads the user's own settings and `CLAUDE.md` files (its sign-in lives there too), so they can change how it answers.
- A message is sent through the standard input of one program run, so a long answer cannot be edited while it is written.
- Pictures cannot be attached to a message yet; Claude looks at the window by itself with `snapshot`.
- `open_image` and `save_image` take any path with a picture extension, so a chat can open and write pictures anywhere the user can.
