# Stickerz

<img src="assets/icon.png" alt="Stickerz icon: a peeled sticker with the bottom-right corner folded over" width="128" height="128">

Stickerz is Nick Greenfield's sticker go/return mod for the reMarkable Paper Pro. It adds a peeled-sticker button on every open document and a back-to-document button on the Stickers notebook, so you can pick a sticker with the native lasso, copy it, return, and paste. This repository is version **1.0**. It has been tested on Nick's Paper Pro, firmware **3.28.0.172**.

There is no picker panel and no one-tap stamp. Copy and paste stay native reMarkable behaviour.

## How it works

1. In any open document, tap the **peeled-sticker** icon at the **top right** (just left of the close X). Stickerz remembers that document and page, then opens the **Stickers** notebook.
2. On a sticker page, use the normal **lasso / selection (capture) tool**, select a sticker, and **Copy**.
3. Tap the **back-to-document** icon at the **top right** of the Stickers notebook (same spot as the sticker icon). You return to where you were.
4. **Paste** the sticker into your document.

The sticker icon is hidden inside the Stickers notebook; the back icon takes its place there.

| Button | Design |
| --- | --- |
| Open Stickers | Peeled sticker: a square with the bottom-right corner folded, fold shaded light grey |
| Back | Left arrow pointing at a lined document page |

Both icons are drawn in QML with plain `Canvas` (black on white for e-ink), 88×88, with `topMargin: 16` and `rightMargin: 120`.

## Requirements

- A reMarkable Paper Pro on firmware **3.28** (tested on **3.28.0.172**), with developer mode on.
- [XOVI](https://github.com/asivery/rm-xovi-extensions) and **qt-resource-rebuilder**, already installed and working (for example via [reManager](https://github.com/asivery/rm-xovi-extensions)).
- A Windows 10 or Windows 11 PC. Over USB, the installer's default address `10.11.99.1` works. Over Wi-Fi, use the tablet's IP.

## Installation (Windows installer)

1. Download `Stickerz-1.0-Setup.zip` from the [Releases](https://github.com/nickgeenfield/Stickerz/releases) page and extract it. The program is `Stickerz-1.0-Setup.exe`.
2. Windows SmartScreen may warn that it does not recognise the app. Choose **More info**, then **Run anyway**.
3. Plug the tablet in by USB (or join the same Wi-Fi) and unlock it.
4. The SSH user is `root`. The password is under **Settings > General > About > Copyrights and licences** (scroll to the bottom). Developer mode is required. The password is case-sensitive and can change after a software update.
5. Run `Stickerz-1.0-Setup.exe`. Enter the IP address (default `10.11.99.1`), the user `root`, and the password. The password is masked as you type and is **never stored**.
6. When the menu appears, choose:
   - **1** — Install or update Stickerz
   - **2** — Uninstall Stickerz
7. After install, in **reManager**: **rebuild the hashtab**, then **restart XOVI / xochitl**. The installer does **not** restart the UI for you.

Upgrading from the older **rm-stickers** v3 / v3.1 packages is supported: `install.sh` moves the old `rm-stickers.qmd` aside (with a backup) so only `stickerz.qmd` loads. Your **Stickers** notebook is kept.

## Manual install

1. Download `Stickerz-1.0.zip` from [Releases](https://github.com/nickgeenfield/Stickerz/releases) and extract it so you have a `Stickerz-1.0` folder containing `install.sh`.
2. Copy it to the tablet:

```sh
scp -r Stickerz-1.0 root@10.11.99.1:/home/root/stickerz
```

3. SSH in and run:

```sh
sh /home/root/stickerz/install.sh
```

4. In **reManager**: rebuild the hashtab, then restart XOVI / xochitl.

### What install.sh does

- Prints the firmware version.
- Backs up `exthome/qt-resource-rebuilder` to `/home/root/stickerz-backup-<time>/` before changing anything.
- Never overwrites a file it did not create (or that does not look like Stickerz / rm-stickers).
- Moves `rm-stickers.qmd` (from v3/v3.1) into the backup so both mods do not load.
- Moves only a leftover v1 `stickerPicker.qmd` (which needs a native `.so`) into that backup.
- Leaves `extensions.d` alone.
- Installs `/home/root/xovi/exthome/qt-resource-rebuilder/stickerz.qmd`.
- Records created paths in `/home/root/xovi/stickerz.manifest`.
- Adds the Stickers notebook only if it is not already there.

## Uninstall

```sh
sh /home/root/stickerz/uninstall.sh
```

Or run the Windows installer and choose **2**.

That removes only what is listed in the Stickerz (and leftover rm-stickers) manifest. It keeps the Stickers notebook unless you pass `--remove-library`. Afterwards, rebuild the hashtab and restart in reManager.

## If it goes wrong

XOVI is tethered. If a hash does not resolve, xochitl can crash-loop and the tablet reboots to stock (no mods).

1. SSH in: `ssh root@10.11.99.1`
2. Run `sh /home/root/stickerz/uninstall.sh`  
   or delete `/home/root/xovi/exthome/qt-resource-rebuilder/stickerz.qmd`  
   (and `rm-stickers.qmd` if it is still present)
3. Rebuild the hashtab and restart XOVI in reManager.

## Repository layout

| Path | Contents |
| --- | --- |
| `src/stickerz.qmd` | The QML mod loaded by qt-resource-rebuilder |
| `library/` | The Stickers notebook (reMarkable document) |
| `install.sh` / `uninstall.sh` | On-tablet install and uninstall |
| `installer/` | Go source for the Windows installer (payload embeds the files above) |
| `assets/` | Icons used on this page and by the installer |

Build outputs (`.exe`, `.zip`, `.syso`) are **not** committed. Download them from [Releases](https://github.com/nickgeenfield/Stickerz/releases).

## Licence

GPL-3.0-or-later. Author: Nick Greenfield. See [LICENSE](LICENSE).

Approaches derived from / credits:

- alefaraci / linkFromSelection (GPL-3.0-only)
- FouzR patterns used in earlier versions (GPL-3.0)
- ingatellent hideDocumentClose / stickers research (**GPL-2.0-or-later**) — credit kept; their SceneAssistant native path is **not** used
- rmitchellscott touchLock / share-button-to-menu research (GPL-3.0-only) — cited for toolbar evidence only

XOVI and qt-resource-rebuilder: [asivery](https://github.com/asivery).

### Evidence / provenance

- **AFFECT `8850026134298937527` / open-doc helper** and **AFFECT `11806562588218124596` / DocumentView overlay**: same targets as alefaraci `linkFromSelection.qmd` (3.28) and rm-stickers v3.1.
- **`open(doc, pageIdx)`** via hashed call `9497240032101537789`: linkFromSelection 3.28.
- **Page jump** via `7713405014271992` + enum path from linkFromSelection 3.28.
- **Library lookup** `233694067165438` / `8399431778896295215`: linkFromSelection 3.28.
- **Close-button id** `476924601591820571`: hideDocumentClose 3.28 (visibility only — not used for INSERT).
- Plain QtQuick types (`Item`, `Rectangle`, `MouseArea`, `Text`, `Timer`, `Canvas`) are unhashed.
