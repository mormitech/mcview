#!/bin/sh
# Makes mormiView show up in Open With and in the application menu of the Linux desktop (per user, no root needed).
# Run it from the folder the tar.gz was extracted to:  ./install.sh      Undo it with:  ./install.sh --remove
set -e
here=$(cd "$(dirname "$0")" && pwd)
apps="${XDG_DATA_HOME:-$HOME/.local/share}/applications"
desktop="$apps/mormiview.desktop"

if [ "$1" = "--remove" ]; then
  rm -f "$desktop" "$apps/mcview.desktop"
else
  mkdir -p "$apps"
  rm -f "$apps/mcview.desktop" # (the entry of the old name, mcView)
  # the desktop file names the program and the icon without a path; point both to where this copy is
  sed -e "s|^Exec=mormiview |Exec=$here/mormiview |" -e "s|^Icon=mormiview|Icon=$here/appicon.png|" "$here/mormiview.desktop" > "$desktop"
fi
command -v update-desktop-database >/dev/null 2>&1 && update-desktop-database "$apps" || true
echo "mormiView: done. Right-click a picture, Open With, mormiView."
