# Stickerz 1.0

QML-only sticker go/return buttons for reMarkable Paper Pro, firmware **3.28.0.172**, with XOVI and qt-resource-rebuilder.

**Author:** Nick Greenfield  
**Version:** 1.0  
**Based on:** rm-stickers v3.1 (confirmed working on Nick's tablet: sticker + back icons and lasso copy).

Stickerz is a new app name for that working mod. The **Stickers** notebook (library of sticker pages) is unchanged.

## What it does

1. In any open document, a **peeled-sticker** button appears at the **top right**, inset left of the usual close (X) area. Tap it: Stickerz remembers your current document and page on a long-lived root object, then opens the **Stickers** notebook (last category page you used, or the first page).
2. In Stickers, use the normal **selection / capture (lasso) tool**, select a sticker, and **Copy**.
3. A **back-to-document** button (left arrow + lined page) appears in the **same top-right spot** whenever Stickers is open. Tap it to return, then **Paste**. If nothing was remembered, it shows a short message — use the native X to close.
4. The sticker button is **hidden inside the Stickers notebook**; Back takes its place there.

There is no picker panel and no one-tap stamp. Copy and paste stay native.

## Icons

Drawn in QML with plain `Canvas` (no new hashes), black on white for e-ink:

| Button | Design (Nick's pick) |
|---|---|
| Open Stickers | **C — Peeled sticker**: square with bottom-right corner folded, fold shaded light grey |
| Back | **4 — Back to document**: left arrow pointing at a lined document page |

Same positions, sizes (88×88) and margins as v3.1 (`topMargin: 16`, `rightMargin: 120`).

## Install on Windows (tablet on USB at 10.11.99.1)

You need: the tablet plugged in by USB (or on Wi-Fi with SSH), and your SSH password from Settings > General > About > Copyrights and licences (bottom).

### Recommended: Windows installer

1. Run `Stickerz-1.0-Setup.exe`. Windows SmartScreen may warn — choose **More info**, then **Run anyway**.
2. Enter the tablet IP (default `10.11.99.1`), username `root`, and the SSH password (never stored).
3. Choose **1** to install or update, or **2** to uninstall.
4. After install, in **reManager**: **rebuild the hashtab**, then **restart XOVI / xochitl**.  
   The installer does **not** restart the UI for you.

Upgrading from rm-stickers **v3 / v3.1** is supported: install.sh moves the old `rm-stickers.qmd` aside (backup first) so only `stickerz.qmd` loads, and clears the old manifest. The Stickers notebook is kept.

### Manual (scp + ssh)

1. Extract `Stickerz-1.0.zip` so you have a `Stickerz-1.0` folder containing `install.sh`.
2. Copy to the tablet:

```
scp -r Stickerz-1.0 root@10.11.99.1:/home/root/stickerz
```

3. SSH in and run:

```
sh /home/root/stickerz/install.sh
```

4. In **reManager**: rebuild the hashtab, then restart XOVI / xochitl.

### What install.sh does

* Prints the firmware version.
* Backs up `exthome/qt-resource-rebuilder` to `/home/root/stickerz-backup-<time>/` before changing anything.
* Never overwrites a file it did not create (or that does not look like Stickerz / rm-stickers).
* Moves `rm-stickers.qmd` (from v3/v3.1) into the backup so both mods do not load.
* Moves only a leftover v1 `stickerPicker.qmd` (needs native `.so`) into that backup.
* Leaves `extensions.d` alone.
* Installs `/home/root/xovi/exthome/qt-resource-rebuilder/stickerz.qmd`.
* Records created paths in `/home/root/xovi/stickerz.manifest`.
* Adds the Stickers notebook only if it is not already there.

## Uninstall

```
sh /home/root/stickerz/uninstall.sh
```

Removes only what is listed in the Stickerz (and leftover rm-stickers) manifest. Keeps the Stickers notebook unless you pass `--remove-library`. Afterwards, rebuild the hashtab and restart in reManager.

## If it goes wrong

XOVI is tethered. If a hash does not resolve, xochitl can crash-loop and the tablet reboots to stock (no mods).

1. SSH in: `ssh root@10.11.99.1`
2. Run `sh /home/root/stickerz/uninstall.sh`  
   or delete `/home/root/xovi/exthome/qt-resource-rebuilder/stickerz.qmd`  
   (and `rm-stickers.qmd` if still present)
3. Rebuild the hashtab and restart XOVI in reManager.

## Evidence / provenance

* **AFFECT `8850026134298937527` / open-doc helper** and **AFFECT `11806562588218124596` / DocumentView overlay**: same targets as alefaraci `research/alefaraci/3.28/linkFromSelection.qmd` and rm-stickers v3.1.
* **`open(doc, pageIdx)`** via hashed call `9497240032101537789`: linkFromSelection 3.28.
* **Page jump** via `7713405014271992` + enum path from linkFromSelection 3.28.
* **Library lookup** `233694067165438` / `8399431778896295215`: linkFromSelection 3.28.
* **Close-button id** `476924601591820571`: hideDocumentClose 3.28 (visibility only — not used for INSERT).
* Plain QtQuick types (`Item`, `Rectangle`, `MouseArea`, `Text`, `Timer`, `Canvas`) are unhashed.

## Licence

GPL-3.0-or-later. Author: Nick Greenfield.

Approaches derived from / credits:

* alefaraci / linkFromSelection (GPL-3.0-only)
* FouzR patterns used in earlier versions (GPL-3.0)
* ingatellent hideDocumentClose / stickers research (**GPL-2.0-or-later**) — credit kept; their SceneAssistant native path is **not** used
* rmitchellscott touchLock / share-button-to-menu research (GPL-3.0-only) — cited for toolbar evidence only

XOVI and qt-resource-rebuilder: asivery.
