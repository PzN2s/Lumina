<p align="center">
  <img src="https://img.shields.io/badge/go-1.26%2B-%2300ADD8?style=for-the-badge&logo=go" alt="Go 1.26+">
  <img src="https://img.shields.io/badge/license-MIT-%23blue?style=for-the-badge" alt="MIT">
  <img src="https://img.shields.io/badge/status-beta-%23orange?style=for-the-badge" alt="Beta">
</p>

<h1 align="center">✦ Lumina ✦</h1>
<p align="center"><b>A TUI system cleaner &amp; monitor for Linux 🐧</b></p>

<p align="center">
  <i>Because clicking through settings panels is for peasants.</i>
</p>

<br>

---

## Install

```bash
git clone https://github.com/PzN2s/Lumina
cd Lumina
go build -o lumina .
sudo ./lumina
```

> Requires **Go 1.26+** and **root privileges** (sudo / pkexec).

---

## Controls

| Key | Action |
|:---:|:---|
| `↑` `↓` | Navigate items |
| `Space` | Toggle selection |
| `Enter` | Clean selected targets |
| `Tab` | Switch panel |
| `q` / `Ctrl+C` | Quit |

---

## Panels

**Cleaner** — Browse system cache directories, see what's inside, select what to nuke. File preview per target with animated scanning.

**Monitor** — Live dashboard with CPU, RAM, swap, disk usage, uptime, load average, GPU info (vendor, VRAM, clock speeds), and top processes sorted by memory.

**Theme** — 15 color schemes. Ocean, Forest, Sunset, Aurora, Lavender, Coral, Dusk, Slate, Rose, Amber, Indigo, Teal, Cherry, Sage, Midnight.

---

## Distro Support

| Distro | Auto-clean | Dependency | Fallback |
|:---|:---:|:---|:---|
| Ubuntu / Debian / Mint / Pop / Zorin / Elementary / Deepin / Raspbian | yes | `apt` (built-in) | — |
| Fedora / CentOS / Rocky / AlmaLinux | yes | `dnf` (built-in) | — |
| Arch / ArchARM / Manjaro / EndeavourOS / CachyOS | yes | `pacman` (built-in) | `paccache` -> `pacman -Sc` |
| Void Linux | yes | `xbps` (built-in) | — |
| Alpine Linux | yes | `apk` (built-in) | — |
| NixOS | yes | `nix-collect-garbage` | — |
| openSUSE Leap / Tumbleweed | yes | `zypper` (built-in) | — |
| Gentoo | yes | `eclean` (from `app-portage/gentoolkit`) | — |

### Optional

```
pciutils -> lspci (GPU detection fallback, works fine without it)
```

---

## How it works

Pulls system data from `/proc`, `/sys`, and `/sys/class/drm`. GPU detection uses the kernel device tree first, falls back to `lspci`. All reads are read-only until you hit Enter in the Cleaner tab — then it runs your distro's package manager cleanup commands alongside removing selected cache dirs.

Built on [`bubbletea`](https://github.com/charmbracelet/bubbletea) and [`lipgloss`](https://github.com/charmbracelet/lipgloss).

---

## License

MIT — do whatever the hell you want with it.
