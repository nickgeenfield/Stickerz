#!/bin/sh
# Stickerz 1.0 installer - upgrades cleanly over rm-stickers v3 / v3.1.
# QML only, no native .so. Run on the tablet as root from the unpacked folder.
set -eu
HERE=$(cd "$(dirname "$0")" && pwd)
XOVI=/home/root/xovi
QMD_DIR=$XOVI/exthome/qt-resource-rebuilder
XO=/home/root/.local/share/remarkable/xochitl
LIB=5d1c4e2a-57ac-4b3a-9f00-000000000000
MANIFEST=$XOVI/stickerz.manifest
OLD_MANIFEST=$XOVI/rm-stickers.manifest
OLD_QMD=$QMD_DIR/rm-stickers.qmd
NEW_QMD=$QMD_DIR/stickerz.qmd
STAMP=$(date +%Y%m%d-%H%M%S)
BACKUP=/home/root/stickerz-backup-$STAMP

fw=$(sed -n 's/^IMG_VERSION="\{0,1\}\([^"]*\)"\{0,1\}$/\1/p' /etc/os-release 2>/dev/null | head -n1)
[ -n "$fw" ] || fw=$(sed -n 's/^REMARKABLE_RELEASE_VERSION=//p' /usr/share/remarkable/update.conf 2>/dev/null | head -n1)
echo "Firmware: ${fw:-could not read /etc/os-release}"
case "$fw" in 3.28.*) ;; *) echo "WARNING: built for 3.28.x only. Press Ctrl-C now to stop, or Enter to continue anyway."; read _ ;; esac

[ -d "$QMD_DIR" ] || { echo "ERROR: $QMD_DIR not found - is XOVI + qt-resource-rebuilder installed?"; exit 1; }
[ -d "$XO" ] || { echo "ERROR: $XO not found"; exit 1; }

# 1. Back up BEFORE changing anything
mkdir -p "$BACKUP"
tar -czf "$BACKUP/qt-resource-rebuilder.tar.gz" -C "$XOVI/exthome" qt-resource-rebuilder
ls -la "$XOVI/extensions.d" > "$BACKUP/extensions.d-listing.txt" 2>&1 || true
[ -f "$OLD_MANIFEST" ] && cp "$OLD_MANIFEST" "$BACKUP/rm-stickers.manifest" || true
[ -f "$MANIFEST" ] && cp "$MANIFEST" "$BACKUP/stickerz.manifest" || true
echo "Backup written to $BACKUP"

# Helper: was this path recorded as ours?
is_ours() {
  f=$1
  grep -qx "$f" "$MANIFEST" 2>/dev/null && return 0
  grep -qx "$f" "$OLD_MANIFEST" 2>/dev/null && return 0
  return 1
}

# Helper: drop a path from a manifest file (if present)
drop_from_manifest() {
  mf=$1; path=$2
  [ -f "$mf" ] || return 0
  grep -vx "$path" "$mf" > "$mf.tmp" 2>/dev/null || true
  if [ -s "$mf.tmp" ]; then mv "$mf.tmp" "$mf"; else rm -f "$mf" "$mf.tmp"; fi
}

# 2. Refuse to overwrite stickerz.qmd if we did not create it
if [ -e "$NEW_QMD" ] && ! is_ours "$NEW_QMD"; then
  # Allow overwrite only if it looks like our own Stickerz file (re-install)
  if grep -q "Stickerz\|stkLayer\|stkBackDoc" "$NEW_QMD" 2>/dev/null; then
    echo "NOTE: replacing existing $NEW_QMD (looks like Stickerz / rm-stickers)."
  else
    echo "ERROR: $NEW_QMD exists but was not installed by Stickerz / rm-stickers. Not touching it."; exit 1
  fi
fi

# 3. Remove previous rm-stickers.qmd so both mods do not load
if [ -e "$OLD_QMD" ]; then
  if is_ours "$OLD_QMD" || grep -q "rm-stickers\|stkLayer\|stkBackDoc\|stkBack" "$OLD_QMD" 2>/dev/null; then
    mv "$OLD_QMD" "$BACKUP/rm-stickers.qmd"
    echo "Moved previous $OLD_QMD to $BACKUP (so only Stickerz loads)."
    drop_from_manifest "$OLD_MANIFEST" "$OLD_QMD"
    drop_from_manifest "$MANIFEST" "$OLD_QMD"
  else
    echo "ERROR: $OLD_QMD exists but was not installed by rm-stickers / Stickerz. Not touching it."; exit 1
  fi
fi

# 4. v1 leftover: stickerPicker.qmd that needs scene-assistant.so
if [ -f "$QMD_DIR/stickerPicker.qmd" ] && grep -q "dk.ingatellent.SceneAssistant" "$QMD_DIR/stickerPicker.qmd"; then
  mv "$QMD_DIR/stickerPicker.qmd" "$BACKUP/stickerPicker.qmd.v1"
  echo "Moved the rm-stickers v1 stickerPicker.qmd to $BACKUP (it needs the native .so)."
fi
if [ -e "$XOVI/extensions.d/scene-assistant.so" ]; then
  echo "NOTE: $XOVI/extensions.d/scene-assistant.so is present. Stickerz does not need it; remove it yourself if you want (this script will not)."
fi

# 5. Install Stickerz qmd
: > "$MANIFEST.new"
cp "$HERE/src/stickerz.qmd" "$NEW_QMD"
echo "$NEW_QMD" >> "$MANIFEST.new"
echo "Installed $NEW_QMD"

# 6. Stickers library notebook (only if not already there; never overwritten)
if [ -e "$XO/$LIB.metadata" ]; then
  echo "Stickers library notebook already on the tablet - left as is."
  # Prefer carrying library lines from old or current manifest
  for mf in "$MANIFEST" "$OLD_MANIFEST"; do
    [ -f "$mf" ] || continue
    grep "$XO/$LIB" "$mf" >> "$MANIFEST.new" 2>/dev/null || true
  done
else
  cp -r "$HERE/library/$LIB" "$XO/$LIB"
  for e in metadata content; do cp "$HERE/library/$LIB.$e" "$XO/$LIB.$e"; echo "$XO/$LIB.$e" >> "$MANIFEST.new"; done
  echo "$XO/$LIB" >> "$MANIFEST.new"
  echo "Added the 'Stickers' library notebook."
fi
sort -u "$MANIFEST.new" > "$MANIFEST"; rm -f "$MANIFEST.new"

# Old rm-stickers.manifest: leave library lines if any remain; drop empty file
if [ -f "$OLD_MANIFEST" ]; then
  grep "^$XO/" "$OLD_MANIFEST" > "$OLD_MANIFEST.tmp" 2>/dev/null || true
  if [ -s "$OLD_MANIFEST.tmp" ]; then
    # Library already tracked in stickerz.manifest — clear the old one to avoid double bookkeeping
    rm -f "$OLD_MANIFEST" "$OLD_MANIFEST.tmp"
    echo "Cleared leftover rm-stickers.manifest (library now tracked by stickerz.manifest)."
  else
    rm -f "$OLD_MANIFEST" "$OLD_MANIFEST.tmp"
  fi
fi

cat <<MSG

Done. Stickerz 1.0 is installed.
Now, in reManager:
  1. Rebuild the hashtab  (or run: $XOVI/rebuild_hashtable)
  2. Restart XOVI / xochitl  (or run: $XOVI/start)
This installer does NOT restart the UI for you.
A peeled-sticker button appears top-right (left of the close area) in documents.
A back-to-document button appears in the same top-right spot when Stickers is open.
If xochitl crash-loops, the tablet reboots to stock (XOVI is tethered) - then run ./uninstall.sh and restart.
MSG
