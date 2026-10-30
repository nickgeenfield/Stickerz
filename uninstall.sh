#!/bin/sh
# Removes only the files listed in the manifest written by Stickerz install.sh.
# Also removes a leftover rm-stickers.qmd if the old manifest still lists it.
set -eu
XOVI=/home/root/xovi
XO=/home/root/.local/share/remarkable/xochitl
QMD_DIR=$XOVI/exthome/qt-resource-rebuilder
MANIFEST=$XOVI/stickerz.manifest
OLD_MANIFEST=$XOVI/rm-stickers.manifest
KEEP_LIB=1; [ "${1:-}" = "--remove-library" ] && KEEP_LIB=0

if [ ! -f "$MANIFEST" ] && [ ! -f "$OLD_MANIFEST" ]; then
  echo "No Stickerz / rm-stickers manifest found. Nothing removed."
  exit 0
fi

remove_listed() {
  mf=$1
  [ -f "$mf" ] || return 0
  while IFS= read -r f; do
    [ -n "$f" ] || continue
    case "$f" in
      $XO/*) [ $KEEP_LIB = 1 ] && { echo "Kept library: $f (use --remove-library to delete)"; continue; } ;;
    esac
    if [ -e "$f" ]; then rm -rf -- "$f"; echo "Removed $f"; else echo "Already gone: $f"; fi
  done < "$mf"
}

remove_listed "$MANIFEST"
# Leftover v3/v3.1 qmd if still present and recorded
if [ -f "$OLD_MANIFEST" ]; then
  remove_listed "$OLD_MANIFEST"
fi
# Belt-and-braces: if old qmd still sits there and looks like ours, remove it
OLD_QMD=$QMD_DIR/rm-stickers.qmd
if [ -f "$OLD_QMD" ] && grep -q "stkLayer\|rm-stickers\|stkBackDoc" "$OLD_QMD" 2>/dev/null; then
  rm -f -- "$OLD_QMD"
  echo "Removed leftover $OLD_QMD"
fi

if [ $KEEP_LIB = 0 ]; then
  rm -f "$MANIFEST" "$OLD_MANIFEST"
else
  for mf in "$MANIFEST" "$OLD_MANIFEST"; do
    [ -f "$mf" ] || continue
    grep "^$XO/" "$mf" > "$mf.tmp" 2>/dev/null || true
    if [ -s "$mf.tmp" ]; then mv "$mf.tmp" "$mf"; else rm -f "$mf" "$mf.tmp"; fi
  done
fi
echo "Now rebuild the hashtab and restart via reManager."
