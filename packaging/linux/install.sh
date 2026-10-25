#!/bin/sh
# Makes mcView show up in Open With and in the application menu of the Linux desktop (per user, no root needed).
# Run it from the folder the tar.gz was extracted to:  ./install.sh      Undo it with:  ./install.sh --remove
set -e
here=$(cd "$(dirname "$0")" && pwd)
apps="${XDG_DATA_HOME:-$HOME/.local/share}/applications"
desktop="$apps/mcview.desktop"

if [ "$1" = "--remove" ]; then
  rm -f "$desktop"
else
  mkdir -p "$apps"
  # the desktop file names the program and the icon without a path; point both to where this copy is
  sed -e "s|^Exec=mcview |Exec=$here/mcview |" -e "s|^Icon=mcview|Icon=$here/appicon.png|" "$here/mcview.desktop" > "$desktop"
fi
command -v update-desktop-database >/dev/null 2>&1 && update-desktop-database "$apps" || true
echo "mcView: done. Right-click a picture, Open With, mcView."
