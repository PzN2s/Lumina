package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Theme ThemeType `yaml:"theme"`
}

func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	configDir := filepath.Join(home, ".config", "lumina")
	_ = os.MkdirAll(configDir, 0755)
	return filepath.Join(configDir, "config.yaml")
}

func loadConfig() Config {
	path := getConfigPath()
	if path == "" {
		return Config{Theme: ThemeOcean}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{Theme: ThemeOcean}
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{Theme: ThemeOcean}
	}

	if _, ok := themes[cfg.Theme]; !ok {
		return Config{Theme: ThemeOcean}
	}

	return cfg
}

func saveConfig(theme ThemeType) {
	path := getConfigPath()
	if path == "" {
		return
	}

	cfg := Config{Theme: theme}
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return
	}

	_ = os.WriteFile(path, data, 0644)
}

type ThemeType string

const (
	ThemeOcean    ThemeType = "ocean"
	ThemeForest   ThemeType = "forest"
	ThemeSunset   ThemeType = "sunset"
	ThemeAurora   ThemeType = "aurora"
	ThemeLavender ThemeType = "lavender"
	ThemeCoral    ThemeType = "coral"
	ThemeDusk     ThemeType = "dusk"
	ThemeSlate    ThemeType = "slate"
	ThemeRose     ThemeType = "rose"
	ThemeAmber    ThemeType = "amber"
	ThemeIndigo   ThemeType = "indigo"
	ThemeTeal     ThemeType = "teal"
	ThemeCherry   ThemeType = "cherry"
	ThemeSage     ThemeType = "sage"
	ThemeMidnight ThemeType = "midnight"
)

type Theme struct {
	name       string
	primary    lipgloss.Color
	secondary  lipgloss.Color
	muted      lipgloss.Color
	success    lipgloss.Color
	warning    lipgloss.Color
	error      lipgloss.Color
	highlight  lipgloss.Color
	dimText    lipgloss.Color
	brightText lipgloss.Color
}

var themes = map[ThemeType]Theme{
	ThemeOcean: {
		name: "Ocean", primary: lipgloss.Color("#0891b2"), secondary: lipgloss.Color("#38bdf8"),
		muted: lipgloss.Color("#164e63"), success: lipgloss.Color("#06b6d4"),
		warning: lipgloss.Color("#22d3ee"), error: lipgloss.Color("#f43f5e"),
		highlight: lipgloss.Color("#7dd3fc"), dimText: lipgloss.Color("#bae6fd"),
		brightText: lipgloss.Color("#f0f9ff"),
	},
	ThemeForest: {
		name: "Forest", primary: lipgloss.Color("#166534"), secondary: lipgloss.Color("#22c55e"),
		muted: lipgloss.Color("#14532d"), success: lipgloss.Color("#4ade80"),
		warning: lipgloss.Color("#86efac"), error: lipgloss.Color("#f97316"),
		highlight: lipgloss.Color("#86efac"), dimText: lipgloss.Color("#bbf7d0"),
		brightText: lipgloss.Color("#f0fdf4"),
	},
	ThemeSunset: {
		name: "Sunset", primary: lipgloss.Color("#ea580c"), secondary: lipgloss.Color("#f97316"),
		muted: lipgloss.Color("#7c2d12"), success: lipgloss.Color("#84cc16"),
		warning: lipgloss.Color("#fbbf24"), error: lipgloss.Color("#dc2626"),
		highlight: lipgloss.Color("#fed7aa"), dimText: lipgloss.Color("#ffedd5"),
		brightText: lipgloss.Color("#fff7ed"),
	},
	ThemeAurora: {
		name: "Aurora", primary: lipgloss.Color("#7c3aed"), secondary: lipgloss.Color("#a855f7"),
		muted: lipgloss.Color("#4c1d95"), success: lipgloss.Color("#06b6d4"),
		warning: lipgloss.Color("#f472b6"), error: lipgloss.Color("#f87171"),
		highlight: lipgloss.Color("#c4b5fd"), dimText: lipgloss.Color("#ddd6fe"),
		brightText: lipgloss.Color("#faf5ff"),
	},
	ThemeLavender: {
		name: "Lavender", primary: lipgloss.Color("#8b5cf6"), secondary: lipgloss.Color("#a78bfa"),
		muted: lipgloss.Color("#5b21b6"), success: lipgloss.Color("#6ee7b7"),
		warning: lipgloss.Color("#c4b5fd"), error: lipgloss.Color("#f472b6"),
		highlight: lipgloss.Color("#ddd6fe"), dimText: lipgloss.Color("#e9d5ff"),
		brightText: lipgloss.Color("#f5f3ff"),
	},
	ThemeCoral: {
		name: "Coral", primary: lipgloss.Color("#e11d48"), secondary: lipgloss.Color("#fb7185"),
		muted: lipgloss.Color("#881337"), success: lipgloss.Color("#4ade80"),
		warning: lipgloss.Color("#fbbf24"), error: lipgloss.Color("#f97316"),
		highlight: lipgloss.Color("#fecdd3"), dimText: lipgloss.Color("#ffe4e6"),
		brightText: lipgloss.Color("#fff1f2"),
	},
	ThemeDusk: {
		name: "Dusk", primary: lipgloss.Color("#6366f1"), secondary: lipgloss.Color("#818cf8"),
		muted: lipgloss.Color("#312e81"), success: lipgloss.Color("#34d399"),
		warning: lipgloss.Color("#c4b5fd"), error: lipgloss.Color("#f87171"),
		highlight: lipgloss.Color("#a5b4fc"), dimText: lipgloss.Color("#c7d2fe"),
		brightText: lipgloss.Color("#e0e7ff"),
	},
	ThemeSlate: {
		name: "Slate", primary: lipgloss.Color("#475569"), secondary: lipgloss.Color("#64748b"),
		muted: lipgloss.Color("#1e293b"), success: lipgloss.Color("#22d3ee"),
		warning: lipgloss.Color("#94a3b8"), error: lipgloss.Color("#f87171"),
		highlight: lipgloss.Color("#cbd5e1"), dimText: lipgloss.Color("#e2e8f0"),
		brightText: lipgloss.Color("#f8fafc"),
	},
	ThemeRose: {
		name: "Rose", primary: lipgloss.Color("#e11d48"), secondary: lipgloss.Color("#f43f5e"),
		muted: lipgloss.Color("#9f1239"), success: lipgloss.Color("#34d399"),
		warning: lipgloss.Color("#fb7185"), error: lipgloss.Color("#fb923c"),
		highlight: lipgloss.Color("#fda4af"), dimText: lipgloss.Color("#ffe4e6"),
		brightText: lipgloss.Color("#fff1f2"),
	},
	ThemeAmber: {
		name: "Amber", primary: lipgloss.Color("#d97706"), secondary: lipgloss.Color("#f59e0b"),
		muted: lipgloss.Color("#78350f"), success: lipgloss.Color("#4ade80"),
		warning: lipgloss.Color("#fbbf24"), error: lipgloss.Color("#ef4444"),
		highlight: lipgloss.Color("#fde68a"), dimText: lipgloss.Color("#fef3c7"),
		brightText: lipgloss.Color("#fffbeb"),
	},
	ThemeIndigo: {
		name: "Indigo", primary: lipgloss.Color("#4338ca"), secondary: lipgloss.Color("#6366f1"),
		muted: lipgloss.Color("#312e81"), success: lipgloss.Color("#22d3ee"),
		warning: lipgloss.Color("#818cf8"), error: lipgloss.Color("#f472b6"),
		highlight: lipgloss.Color("#c7d2fe"), dimText: lipgloss.Color("#e0e7ff"),
		brightText: lipgloss.Color("#eef2ff"),
	},
	ThemeTeal: {
		name: "Teal", primary: lipgloss.Color("#0d9488"), secondary: lipgloss.Color("#14b8a6"),
		muted: lipgloss.Color("#134e4a"), success: lipgloss.Color("#5eead4"),
		warning: lipgloss.Color("#99f6e4"), error: lipgloss.Color("#f87171"),
		highlight: lipgloss.Color("#5eead4"), dimText: lipgloss.Color("#ccfbf1"),
		brightText: lipgloss.Color("#f0fdfa"),
	},
	ThemeCherry: {
		name: "Cherry", primary: lipgloss.Color("#be123c"), secondary: lipgloss.Color("#e11d48"),
		muted: lipgloss.Color("#9f1239"), success: lipgloss.Color("#4ade80"),
		warning: lipgloss.Color("#fb7185"), error: lipgloss.Color("#fb923c"),
		highlight: lipgloss.Color("#fda4af"), dimText: lipgloss.Color("#ffe4e6"),
		brightText: lipgloss.Color("#fff1f2"),
	},
	ThemeSage: {
		name: "Sage", primary: lipgloss.Color("#4d7c0f"), secondary: lipgloss.Color("#65a30d"),
		muted: lipgloss.Color("#365314"), success: lipgloss.Color("#84cc16"),
		warning: lipgloss.Color("#a3e635"), error: lipgloss.Color("#f87171"),
		highlight: lipgloss.Color("#bef264"), dimText: lipgloss.Color("#d9f99d"),
		brightText: lipgloss.Color("#f7fdf5"),
	},
	ThemeMidnight: {
		name: "Midnight", primary: lipgloss.Color("#1e1b4b"), secondary: lipgloss.Color("#312e81"),
		muted: lipgloss.Color("#0f0a29"), success: lipgloss.Color("#06b6d4"),
		warning: lipgloss.Color("#a78bfa"), error: lipgloss.Color("#f472b6"),
		highlight: lipgloss.Color("#6366f1"), dimText: lipgloss.Color("#818cf8"),
		brightText: lipgloss.Color("#e0e7ff"),
	},
}

var themeList = []ThemeType{
	ThemeOcean, ThemeForest, ThemeSunset, ThemeAurora, ThemeLavender,
	ThemeCoral, ThemeDusk, ThemeSlate, ThemeRose, ThemeAmber,
	ThemeIndigo, ThemeTeal, ThemeCherry, ThemeSage, ThemeMidnight,
}

var currentTheme = ThemeOcean

func applyTheme(t ThemeType) Theme {
	return themes[t]
}

func themeHeader(t Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.brightText).
		Background(t.primary).
		Padding(0, 2).
		Bold(true)
}

func themeBorder(t Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.primary).
		Padding(1, 2)
}

func themePath(t Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.muted).
		Italic(true)
}

func themeFolder(t Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.secondary).
		Bold(true)
}

func themeFile(t Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.dimText)
}

func themeGlow(t Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.highlight).
		Bold(true)
}

func themeTitle(t Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.warning).
		Bold(true)
}

func themeTabActive(t Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(t.primary).
		Foreground(t.brightText).
		Bold(true).
		Padding(0, 1)
}

func themeTabMuted(t Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.dimText).
		Padding(0, 1)
}

func themeBarLow(t Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.success)
}

type fileEntry struct {
	name  string
	isDir bool
	size  int64
}

type target struct {
	label    string
	path     string
	selected bool
	bytes    int64
	safe     bool
}

type procInfo struct {
	pid     int
	name    string
	state   string
	cpu     float64
	memMB   float64
	memPct  float64
	command string
}

type gpuInfo struct {
	name       string
	vendor     string
	driver     string
	vramMB     uint64
	curFreqMHz uint64
	maxFreqMHz uint64
}

type model struct {
	spinner      spinner.Model
	progress     progress.Model
	targets      []*target
	index        int
	state        string
	confirming   bool
	osInfo       string
	osIcon       string
	progressPct  float64
	fileEntries  []fileEntry
	fileCount    int
	scanDone     bool
	scanPending  bool
	sizing       bool
	countdown    int
	width        int
	height       int
	tab          int
	cpuPercent   float64
	prevCPUTotal uint64
	prevCPUIdle  uint64
	memTotal     uint64
	memUsed      uint64
	memAvail     uint64
	memPercent   float64
	swapTotal    uint64
	swapUsed     uint64
	swapPercent  float64
	uptime       uint64
	loadAvg      []float64
	hostName     string
	kernelVer    string
	numCPUs      int
	procs        []procInfo
	diskTotal    uint64
	diskUsed     uint64
	diskPct      float64
	gpus         []gpuInfo
	mu           sync.Mutex
	theme        ThemeType
	themeIndex   int
	themeTab     int
	cleanErrors  []string
	updateStatus string
	updateOutput string
}

func detectOS() (name, icon string) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "Linux", "🐧"
	}
	defer file.Close()

	var prettyName, id string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			prettyName = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
		}
		if strings.HasPrefix(line, "ID=") {
			id = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		}
	}

	if prettyName != "" {
		name = prettyName
	} else {
		name = "Linux"
	}

	switch strings.ToLower(id) {
	case "nixos":
		icon = ""
	case "ubuntu", "linuxmint", "pop":
		icon = ""
	case "debian", "raspbian":
		icon = ""
	case "fedora", "centos", "rocky", "almalinux":
		icon = ""
	case "arch", "archarm", "manjaro", "endeavouros", "cachyos":
		icon = ""
	case "opensuse", "opensuse-leap", "opensuse-tumbleweed":
		icon = ""
	case "void":
		icon = ""
	case "gentoo":
		icon = ""
	case "alpine":
		icon = ""
	case "elementary":
		icon = ""
	case "solus":
		icon = ""
	case "deepin":
		icon = ""
	default:
		icon = "🐧"
	}

	return name, icon
}

func buildTargets(osID string) []*target {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/tmp"
	}
	var targets []*target

	targets = append(targets,
		&target{label: "User Cache", path: filepath.Join(home, ".cache"), selected: false, safe: true},
		&target{label: "Media Thumbs", path: filepath.Join(home, ".cache/thumbnails"), selected: false, safe: true},
		&target{label: "User Trash", path: filepath.Join(home, ".local/share/Trash"), selected: false, safe: true},
	)

	switch strings.ToLower(osID) {
	case "ubuntu", "debian", "linuxmint", "pop", "zorin", "elementary", "deepin", "raspbian":
		targets = append(targets,
			&target{label: "APT Cache", path: "/var/cache/apt/archives", selected: false, safe: true},
		)
	case "fedora", "centos", "rocky", "almalinux":
		targets = append(targets,
			&target{label: "DNF Cache", path: "/var/cache/dnf", selected: false, safe: true},
		)
	case "arch", "archarm", "manjaro", "endeavouros", "cachyos":
		targets = append(targets,
			&target{label: "Pacman Cache", path: "/var/cache/pacman/pkg", selected: false, safe: true},
		)
	case "void":
		targets = append(targets,
			&target{label: "Void Cache", path: "/var/cache/xbps", selected: false, safe: true},
		)
	case "alpine":
		targets = append(targets,
			&target{label: "APK Cache", path: "/var/cache/apk", selected: false, safe: true},
		)
	case "gentoo":
		targets = append(targets,
			&target{label: "Distfiles", path: "/var/cache/distfiles", selected: false, safe: true},
		)
	case "nixos":
		targets = append(targets,
			&target{label: "Nix GC", path: "/nix/store", selected: false, safe: true},
		)
	case "opensuse", "opensuse-leap", "opensuse-tumbleweed":
		targets = append(targets,
			&target{label: "Zypper Cache", path: "/var/cache/zypp/packages", selected: false, safe: true},
		)
	}

	if _, err := os.Stat("/tmp"); err == nil {
		targets = append(targets,
			&target{label: "System Temp", path: "/tmp", selected: false, safe: true},
		)
	}

	return targets
}

func diskUsage(p string) int64 {
	var total int64
	_ = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		info, err := d.Info()
		if err != nil || (info.Mode()&os.ModeSymlink != 0) {
			return nil
		}
		if !d.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit && exp < 3; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGT"[exp])
}

func getOSID() string {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			return strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		}
	}
	return ""
}

func getCPUUsage() (totalTicks, idleTicks uint64, nCPUs int) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, 1
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var line string
	if !scanner.Scan() {
		return 0, 0, 1
	}
	line = scanner.Text()

	for scanner.Scan() {
		l := scanner.Text()
		if strings.HasPrefix(l, "cpu") && !strings.HasPrefix(l, "cpu ") {
			nCPUs++
		}
	}
	if nCPUs == 0 {
		nCPUs = 1
	}

	fields := strings.Fields(line)
	if len(fields) < 5 {
		return 0, 0, nCPUs
	}

	var total uint64
	for i := 1; i < len(fields); i++ {
		v, err := strconv.ParseUint(fields[i], 10, 64)
		if err == nil {
			total += v
		}
	}

	idle := uint64(0)
	if len(fields) > 4 {
		v, _ := strconv.ParseUint(fields[4], 10, 64)
		idle += v
	}
	if len(fields) > 5 {
		v, _ := strconv.ParseUint(fields[5], 10, 64)
		idle += v
	}

	return total, idle, nCPUs
}

func calcCPUPercent(prevTotal, prevIdle uint64) float64 {
	newTotal, newIdle, _ := getCPUUsage()

	if newTotal == prevTotal {
		newTotal, newIdle, _ = getCPUUsage()
	}

	dt := newTotal - prevTotal
	di := newIdle - prevIdle

	if dt == 0 {
		return 0
	}

	pct := float64(dt-di) / float64(dt) * 100
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}

	return pct
}

func getMemInfo() (totalMB, usedMB, availMB uint64) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, 0
	}
	defer file.Close()

	var memTotal, memFree, memAvail, buffers, cached, sreclaimable uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		v, _ := strconv.ParseUint(fields[1], 10, 64)
		switch fields[0] {
		case "MemTotal:":
			memTotal = v
		case "MemFree:":
			memFree = v
		case "MemAvailable:":
			memAvail = v
		case "Buffers:":
			buffers = v
		case "Cached:":
			cached = v
		case "SReclaimable:":
			sreclaimable = v
		}
	}

	if memAvail > 0 {
		usedMB = (memTotal - memAvail) / 1024
	} else {
		usedMB = (memTotal - memFree - buffers - cached - sreclaimable) / 1024
	}

	return memTotal / 1024, usedMB, memAvail / 1024
}

func getSwapInfo() (totalMB, usedMB uint64) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	defer file.Close()

	var swapTotal, swapFree uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		v, _ := strconv.ParseUint(fields[1], 10, 64)
		switch fields[0] {
		case "SwapTotal:":
			swapTotal = v
		case "SwapFree:":
			swapFree = v
		}
	}
	usedMB = (swapTotal - swapFree) / 1024
	return swapTotal / 1024, usedMB
}

func getDiskUsage() (totalGB, usedGB uint64, pct float64) {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		return 0, 0, 0
	}
	totalBytes := stat.Blocks * uint64(stat.Bsize)
	availBytes := stat.Bavail * uint64(stat.Bsize)
	usedBytes := totalBytes - availBytes

	totalGB = totalBytes / 1024 / 1024 / 1024
	usedGB = usedBytes / 1024 / 1024 / 1024
	if totalBytes > 0 {
		pct = float64(usedBytes) / float64(totalBytes) * 100
	}
	return
}

var gpuVendorMap = map[string]string{
	"0x8086": "Intel",
	"0x1002": "AMD",
	"0x10de": "NVIDIA",
	"0x1234": "QEMU",
	"0x1af4": "Virtio",
	"0x15ad": "VMware",
}

func detectGPUName(vendor, pciID string) string {
	device := ""
	if parts := strings.Split(pciID, ":"); len(parts) > 1 {
		device = "0x" + parts[1]
	}

	intelMap := map[string]string{
		"0x9a49": "UHD Graphics G4",
		"0x9a40": "UHD Graphics",
		"0x9a42": "UHD Graphics",
		"0x9a44": "UHD Graphics",
		"0x9a48": "Iris Xe Graphics",
		"0x9a50": "Iris Xe Graphics",
		"0x9a59": "Iris Xe Graphics",
		"0x9a60": "Iris Xe Graphics",
		"0x9bc8": "CometLake-S GT2 UHD",
		"0x3e92": "UHD Graphics 630",
		"0x3ea5": "Iris Plus G7",
		"0x3ea0": "UHD Graphics 610",
		"0x46a6": "Alder Lake GT2",
		"0x4680": "Alder Lake GT1",
		"0x4690": "Alder Lake GT2",
		"0x4693": "Iris Xe Aderlake",
		"0xa780": "Raptor Lake GT1",
	}

	amdMap := map[string]string{
		"0x73ff": "RX 6800 XT",
		"0x73a2": "RX 6900 XT",
		"0x747c": "RX 6700 XT",
	}

	nvidiaMap := map[string]string{
		"0x2484": "RTX 4090",
		"0x2204": "RTX 3080",
		"0x2206": "RTX 3070",
	}

	var lookup map[string]string
	switch vendor {
	case "0x1002":
		lookup = amdMap
	case "0x10de":
		lookup = nvidiaMap
	case "0x8086":
		lookup = intelMap
	default:
		lookup = nil
	}

	vendorName := gpuVendorMap[vendor]
	if vendorName == "" {
		vendorName = "Unknown"
	}

	if lookup != nil {
		if name, ok := lookup[device]; ok {
			return vendorName + " " + name
		}
	}

	return vendorName + " Device " + device
}

func getGPUs() []gpuInfo {
	var gpus []gpuInfo

	entries, err := os.ReadDir("/sys/class/drm")
	if err != nil {
		return gpus
	}

	seen := make(map[string]bool)

	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "card") || strings.Contains(e.Name(), "-") {
			continue
		}
		cardName := e.Name()
		if seen[cardName] {
			continue
		}
		seen[cardName] = true

		cardPath := filepath.Join("/sys/class/drm", cardName)
		devPath := filepath.Join(cardPath, "device")

		vendorRaw, err := os.ReadFile(filepath.Join(devPath, "vendor"))
		if err != nil {
			continue
		}
		vendor := strings.TrimSpace(string(vendorRaw))

		driverLink, err := os.Readlink(filepath.Join(devPath, "driver"))
		driver := "unknown"
		if err == nil {
			driver = filepath.Base(driverLink)
		}

		ueventData, err := os.ReadFile(filepath.Join(devPath, "uevent"))
		pciID := ""
		if err == nil {
			for _, line := range strings.Split(string(ueventData), "\n") {
				if strings.HasPrefix(line, "PCI_ID=") {
					pciID = strings.TrimPrefix(line, "PCI_ID=")
					break
				}
			}
		}

		name := detectGPUName(vendor, pciID)

		var vramMB uint64

		amdVRAM, err := os.ReadFile(filepath.Join(devPath, "mem_info_vram_total"))
		if err == nil {
			vBytes, _ := strconv.ParseUint(strings.TrimSpace(string(amdVRAM)), 10, 64)
			vramMB = vBytes / 1024 / 1024
		}

		if vramMB == 0 {
			nvInfo, err := os.ReadDir("/proc/driver/nvidia/gpus")
			if err == nil {
				for _, gpuDir := range nvInfo {
					memFile := filepath.Join("/proc/driver/nvidia/gpus", gpuDir.Name(), "memory")
					memData, err := os.ReadFile(memFile)
					if err == nil {
						for _, line := range strings.Split(string(memData), "\n") {
							if strings.Contains(line, "Total") && strings.Contains(line, "MiB") {
								fields := strings.Fields(line)
								for i, f := range fields {
									if f == "Total" && i+1 < len(fields) {
										v, _ := strconv.ParseUint(fields[i+1], 10, 64)
										vramMB = v
										break
									}
								}
							}
						}
					}
				}
			}
		}

		gpus = append(gpus, gpuInfo{
			name:   name,
			vendor: gpuVendorMap[vendor],
			driver: driver,
			vramMB: vramMB,
		})

		if vendor == "0x8086" {
			curFreqData, _ := os.ReadFile(filepath.Join(cardPath, "gt_act_freq_mhz"))
			if curFreqData != nil {
				gpus[len(gpus)-1].curFreqMHz, _ = strconv.ParseUint(strings.TrimSpace(string(curFreqData)), 10, 64)
			}
			maxFreqData, _ := os.ReadFile(filepath.Join(cardPath, "gt_max_freq_mhz"))
			if maxFreqData != nil {
				gpus[len(gpus)-1].maxFreqMHz, _ = strconv.ParseUint(strings.TrimSpace(string(maxFreqData)), 10, 64)
			}
		}
	}

	if len(gpus) == 0 && commandExists("lspci") {
		out, err := exec.Command("lspci").Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				lower := strings.ToLower(line)
				if strings.Contains(lower, "vga") || strings.Contains(lower, "3d") || strings.Contains(lower, "display") {
					name := "Unknown GPU"
					if idx := strings.Index(line, ": "); idx != -1 {
						name = strings.TrimSpace(line[idx+2:])
					}
					driver := "unknown"
					if strings.Contains(lower, "nvidia") {
						driver = "nvidia"
					} else if strings.Contains(lower, "radeon") || strings.Contains(lower, "amdgpu") {
						driver = "amdgpu"
					} else if strings.Contains(lower, "intel") {
						driver = "i915"
					}
					gpus = append(gpus, gpuInfo{name: name, driver: driver})
				}
			}
		}
	}

	return gpus
}

func getProcesses(memTotalMB uint64) []procInfo {
	pidRe := regexp.MustCompile(`^\d+$`)
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	var procs []procInfo
	pageSize := uint64(os.Getpagesize())

	uptimeSec := uint64(0)
	if data, err := os.ReadFile("/proc/uptime"); err == nil {
		flds := strings.Fields(strings.TrimSpace(string(data)))
		if len(flds) > 0 {
			f, _ := strconv.ParseFloat(flds[0], 64)
			uptimeSec = uint64(f)
		}
	}

	clkTck := uint64(100)

	for _, e := range entries {
		if !pidRe.MatchString(e.Name()) {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}

		statPath := filepath.Join("/proc", e.Name(), "stat")
		data, err := os.ReadFile(statPath)
		if err != nil {
			continue
		}

		statStr := string(data)
		openIdx := strings.Index(statStr, "(")
		closeIdx := strings.LastIndex(statStr, ")")
		if openIdx == -1 || closeIdx == -1 || closeIdx <= openIdx {
			continue
		}

		name := statStr[openIdx+1 : closeIdx]
		fields := strings.Fields(statStr[closeIdx+2:])
		if len(fields) < 22 {
			continue
		}

		state := fields[0]
		utime, _ := strconv.ParseUint(fields[11], 10, 64)
		stime, _ := strconv.ParseUint(fields[12], 10, 64)
		rss, _ := strconv.ParseUint(fields[21], 10, 64)
		starttime, _ := strconv.ParseUint(fields[19], 10, 64)

		totalTime := utime + stime
		seconds := uptimeSec - (starttime / clkTck)
		var cpuPct float64
		if seconds > 0 {
			cpuPct = float64(totalTime) / float64(clkTck) / float64(seconds) * 100
		}
		if cpuPct > 100 {
			cpuPct = 100
		}

		memMB := float64(rss*pageSize) / 1024 / 1024
		var memPct float64
		if memTotalMB > 0 {
			memPct = memMB / float64(memTotalMB) * 100
		}

		cmdPath := filepath.Join("/proc", e.Name(), "cmdline")
		cmdData, err := os.ReadFile(cmdPath)
		cmd := name
		if err == nil && len(cmdData) > 0 {
			parts := strings.Split(strings.TrimRight(string(cmdData), "\x00"), "\x00")
			if len(parts) > 0 {
				if base := filepath.Base(parts[0]); base != "" {
					cmd = base
				} else {
					cmd = name
				}
			}
		}

		procs = append(procs, procInfo{
			pid:     pid,
			name:    name,
			state:   state,
			cpu:     cpuPct,
			memMB:   memMB,
			memPct:  memPct,
			command: cmd,
		})
	}

	sort.Slice(procs, func(i, j int) bool {
		return procs[i].memMB > procs[j].memMB
	})

	if len(procs) > 12 {
		procs = procs[:12]
	}

	return procs
}

func getUptime() uint64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(strings.TrimSpace(string(data)))
	if len(fields) < 1 {
		return 0
	}
	f, _ := strconv.ParseFloat(fields[0], 64)
	return uint64(f)
}

func getLoadAvg() []float64 {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return []float64{0, 0, 0}
	}
	fields := strings.Fields(strings.TrimSpace(string(data)))
	if len(fields) < 3 {
		return []float64{0, 0, 0}
	}
	l1, _ := strconv.ParseFloat(fields[0], 64)
	l5, _ := strconv.ParseFloat(fields[1], 64)
	l15, _ := strconv.ParseFloat(fields[2], 64)
	return []float64{l1, l5, l15}
}

func formatUptime(seconds uint64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	mins := (seconds % 3600) / 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm %ds", mins, seconds%60)
}

func getSysInfo(m *model) {
	h, err := os.Hostname()
	if err != nil {
		m.hostName = "unknown"
	} else {
		m.hostName = h
	}

	kv, err := os.ReadFile("/proc/version")
	if err == nil {
		flds := strings.Fields(string(kv))
		if len(flds) >= 3 {
			m.kernelVer = flds[2]
		}
	}

	_, _, nCPUs := getCPUUsage()
	m.numCPUs = nCPUs
}

type sysMonitorMsg struct{}

func startMonitor() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg { return sysMonitorMsg{} })
}

type scanMsg struct {
	entries []fileEntry
}

type animTickMsg struct{}

type cleanDoneMsg struct {
	errors []string
}

type updateMsg struct {
	success bool
	output  string
}

type sizeMsg struct {
	index int
	size  int64
}

type sizesDoneMsg struct{}

func scanTarget(path string) tea.Cmd {
	return func() tea.Msg {
		if _, err := os.Stat(path); err != nil {
			return scanMsg{entries: nil}
		}

		var entries []fileEntry
		baseDepth := len(strings.Split(path, string(os.PathSeparator)))

		_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			info, err := d.Info()
			if err != nil || (info.Mode()&os.ModeSymlink != 0) {
				return nil
			}

			parts := strings.Split(p, string(os.PathSeparator))
			depth := len(parts) - baseDepth
			if depth > 3 {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			entries = append(entries, fileEntry{
				name:  d.Name(),
				isDir: d.IsDir(),
				size:  info.Size(),
			})

			if len(entries) >= 40 {
				return filepath.SkipDir
			}
			return nil
		})

		return scanMsg{entries: entries}
	}
}

func calcSizes(targets []*target) tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(targets)+1)
	for i, t := range targets {
		idx := i
		p := t.path
		cmds = append(cmds, func() tea.Msg {
			if _, err := os.Stat(p); err == nil {
				return sizeMsg{index: idx, size: diskUsage(p)}
			}
			return sizeMsg{index: idx, size: 0}
		})
	}
	cmds = append(cmds, func() tea.Msg { return sizesDoneMsg{} })
	return tea.Batch(cmds...)
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		scanTarget(m.targets[m.index].path),
		calcSizes(m.targets),
		startMonitor(),
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case scanMsg:
		m.fileEntries = msg.entries
		m.fileCount = 0
		m.scanDone = false
		m.scanPending = false
		return m, animTick()
	case animTickMsg:
		if !m.scanDone && m.fileCount < len(m.fileEntries) {
			m.fileCount += 1
			if m.fileCount >= len(m.fileEntries) {
				m.fileCount = len(m.fileEntries)
				m.scanDone = true
			}
			return m, animTick()
		}
		m.scanDone = true
		return m, nil
	case sizeMsg:
		m.mu.Lock()
		m.targets[msg.index].bytes = msg.size
		m.mu.Unlock()
		return m, nil
	case sizesDoneMsg:
		m.mu.Lock()
		m.sizing = false
		m.mu.Unlock()
		return m, nil
	case tea.KeyMsg:
		if m.updateStatus == "done" || m.updateStatus == "error" {
			m.updateStatus = ""
			m.updateOutput = ""
		}
		if m.confirming {
			switch msg.String() {
			case "y", "Y":
				m.confirming = false
				m.state = "running"
				return m, tick()
			case "n", "N", "esc":
				m.confirming = false
				return m, nil
			}
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.tab = (m.tab + 1) % 3
			return m, nil
		case "up", "k":
			if m.tab == 0 {
				if m.index > 0 {
					m.index--
					if !m.sizing {
						m.scanPending = true
						return m, scanTarget(m.targets[m.index].path)
					}
				}
			} else if m.tab == 2 {
				if m.themeIndex > 0 {
					m.themeIndex--
					m.theme = themeList[m.themeIndex]
					saveConfig(m.theme)
					t := themes[m.theme]
					m.spinner.Style = lipgloss.NewStyle().Foreground(t.primary)
					m.progress = progress.New(progress.WithGradient(string(t.primary), string(t.secondary)))
				}
			}
			return m, nil
		case "down", "j":
			if m.tab == 0 {
				if m.index < len(m.targets)-1 {
					m.index++
					if !m.sizing {
						m.scanPending = true
						return m, scanTarget(m.targets[m.index].path)
					}
				}
			} else if m.tab == 2 {
				if m.themeIndex < len(themeList)-1 {
					m.themeIndex++
					m.theme = themeList[m.themeIndex]
					saveConfig(m.theme)
					t := themes[m.theme]
					m.spinner.Style = lipgloss.NewStyle().Foreground(t.primary)
					m.progress = progress.New(progress.WithGradient(string(t.primary), string(t.secondary)))
				}
			}
			return m, nil
		case " ":
			if m.tab == 0 {
				m.targets[m.index].selected = !m.targets[m.index].selected
			}
		case "enter":
			if m.tab == 0 && m.state == "ready" {
				m.confirming = true
				return m, nil
			}
		case "u":
			if m.state != "running" && m.state != "cleaning" && m.state != "updating" {
				m.updateStatus = "updating"
				m.state = "updating"
				return m, m.update()
			}
		}
	case updateMsg:
		m.updateStatus = "done"
		m.updateOutput = msg.output
		if !msg.success {
			m.updateStatus = "error"
		}
		m.state = "ready"
		return m, nil
	case tickMsg:
		if m.state == "running" {
			m.progressPct += 0.05
			if m.progressPct >= 1.0 {
				m.state = "cleaning"
				return m, m.clean()
			}
			return m, tick()
		}
	case cleanDoneMsg:
		m.cleanErrors = msg.errors
		m.state = "done"
		m.countdown = 5
		return m, startCountdown()
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case countdownMsg:
		m.countdown--
		if m.countdown <= 0 {
			return m, tea.Quit
		}
		return m, startCountdown()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case sysMonitorMsg:
		m.mu.Lock()
		cpuPct := calcCPUPercent(m.prevCPUTotal, m.prevCPUIdle)
		m.cpuPercent = cpuPct
		t, i, _ := getCPUUsage()
		m.prevCPUTotal = t
		m.prevCPUIdle = i

		mt, mu, ma := getMemInfo()
		m.memTotal = mt
		m.memUsed = mu
		m.memAvail = ma
		if mt > 0 {
			m.memPercent = float64(mu) / float64(mt) * 100
		}

		st, su := getSwapInfo()
		m.swapTotal = st
		m.swapUsed = su
		if st > 0 {
			m.swapPercent = float64(su) / float64(st) * 100
		}

		dt, du, dp := getDiskUsage()
		m.diskTotal = dt
		m.diskUsed = du
		m.diskPct = dp

		m.uptime = getUptime()
		m.loadAvg = getLoadAvg()
		m.procs = getProcesses(mt)
		m.gpus = getGPUs()
		m.mu.Unlock()
		return m, startMonitor()
	}
	return m, nil
}

func (m *model) View() string {
	if m.width == 0 {
		return "\n  Initializing..."
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	t := applyTheme(m.theme)

	header := fmt.Sprintf("%s  %s %s | %s | %s / %s / %s\n",
		themeHeader(t).Render(" LUMINA INSPECTOR "),
		m.osIcon, m.osInfo,
		lipgloss.NewStyle().Foreground(themeTitle(t).GetForeground()).Bold(true).Render("DEV : Reham"),
		tabStyle(m.tab == 0, t).Render(" Cleaner "),
		tabStyle(m.tab == 1, t).Render(" Monitor "),
		tabStyle(m.tab == 2, t).Render(" Theme "))

	var content string
	if m.updateStatus == "updating" {
		dialogContent := fmt.Sprintf(
			"%s\n\n%s\n\nUpdating Lumina...\nPlease wait.",
			themeTitle(t).Render("⚠ Updating"),
			m.spinner.View(),
		)
		dialog := lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(t.primary).
			Padding(2, 4).
			Render(dialogContent)
		content = dialog
	} else if m.updateStatus == "done" || m.updateStatus == "error" {
		title := "✔ Update Complete"
		col := t.success
		if m.updateStatus == "error" {
			title = "✘ Update Failed"
			col = t.error
		}
		dialogContent := fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			lipgloss.NewStyle().Foreground(col).Bold(true).Render(title),
			m.updateOutput,
			lipgloss.NewStyle().Foreground(t.muted).Render("Press any key to continue"),
		)
		dialog := lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(col).
			Padding(1, 2).
			Render(dialogContent)
		content = dialog
	} else if m.tab == 0 {
		if m.confirming {
			dialogContent := fmt.Sprintf(
				"%s\n\nAre you sure you want to clean\nthe selected targets?\nThis action cannot be undone.\n\n%s  %s",
				themeTitle(t).Render("⚠ Confirm Clean"),
				lipgloss.NewStyle().Foreground(t.success).Render("[Y] Yes"),
				lipgloss.NewStyle().Foreground(t.error).Render("[N] No"),
			)
			dialog := lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(t.warning).
				Padding(1, 3).
				Render(dialogContent)
			content = dialog
		} else {
			left := m.renderLeft(t)
			right := m.renderRight(t)
			split := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
			content = header + "\n" + split
		}
	} else if m.tab == 1 {
		content = header + "\n" + m.renderMonitor(t)
	} else {
		content = header + "\n" + m.renderTheme(t)
	}

	bordered := themeBorder(t).Render(content)

	lines := strings.Split(bordered, "\n")
	w := lipgloss.Width(bordered)
	h := len(lines)

	hPad := (m.width - w) / 2
	if hPad < 0 {
		hPad = 0
	}
	vPad := (m.height - h) / 2
	if vPad < 0 {
		vPad = 0
	}

	var b strings.Builder
	for row := 0; row < m.height; row++ {
		if row >= vPad && row < vPad+len(lines) {
			line := lines[row-vPad]
			b.WriteString(strings.Repeat(" ", hPad))
			b.WriteString(line)
			lw := hPad + lipgloss.Width(line)
			if m.width > lw {
				b.WriteString(strings.Repeat(" ", m.width-lw))
			}
		} else {
			b.WriteString(strings.Repeat(" ", m.width))
		}
		if row < m.height-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}

func tabStyle(active bool, t Theme) lipgloss.Style {
	if active {
		return themeTabActive(t)
	}
	return themeTabMuted(t)
}

func (m *model) renderLeft(t Theme) string {
	var s string
	if m.sizing {
		s = fmt.Sprintf("%s Scanning...\n\n", m.spinner.View())
	} else {
		s = "Select to purge (Space/Enter):\n\n"
	}

	for i, tg := range m.targets {
		ptr := "  "
		if m.index == i {
			ptr = "> "
		}
		box := "[ ]"
		if tg.selected {
			box = "[x]"
		}
		sz := formatBytes(tg.bytes)
		line := fmt.Sprintf("%s%s %-16s %s", ptr, box, tg.label, sz)
		if m.index == i {
			s += lipgloss.NewStyle().Foreground(themeTitle(t).GetForeground()).Render(line) + "\n"
			s += "    " + themePath(t).Render("↳ "+tg.path) + "\n"
		} else {
			s += line + "\n"
		}
	}

	if m.state == "running" {
		s = fmt.Sprintf("%s Cleaning...\n\n%s",
			m.spinner.View(), m.progress.ViewAs(m.progressPct))
	}
	if m.state == "cleaning" {
		s = fmt.Sprintf("%s Running cleaners...\n\n%s",
			m.spinner.View(), m.progress.ViewAs(1.0))
	}
	if m.state == "done" {
		s = themeTitle(t).Render("✔ System Cleaned!") + "\n"
		for _, err := range m.cleanErrors {
			s += lipgloss.NewStyle().Foreground(t.error).Render("  ⚠ "+err) + "\n"
		}
		s += fmt.Sprintf("\n   Exiting in %d...\n\n", m.countdown)
		s += lipgloss.NewStyle().Foreground(themeTitle(t).GetForeground()).Italic(true).Render("   byeee")
	}

	return lipgloss.NewStyle().Width(32).Render(s)
}

func (m *model) renderRight(t Theme) string {
	w := 48

	if m.scanPending {
		return lipgloss.NewStyle().Width(w).Height(16).Padding(0, 0, 0, 1).
			Border(lipgloss.ThickBorder()).BorderForeground(themeTitle(t).GetForeground()).
			Render("\n\n  " + m.spinner.View() + " scanning files...")
	}

	if len(m.fileEntries) == 0 {
		return lipgloss.NewStyle().Width(w).Height(16).Padding(0, 0, 0, 1).
			Border(lipgloss.ThickBorder()).BorderForeground(themeTitle(t).GetForeground()).
			Render("\n\n  No files found")
	}

	var dirs []string
	var files []string
	for _, e := range m.fileEntries {
		if e.isDir {
			dirs = append(dirs, e.name)
		} else {
			files = append(files, e.name)
		}
	}

	var content string

	dirsTitle := themeTitle(t).Render(" Folders ") + " " + strings.Repeat("─", 30)
	content += dirsTitle + "\n"

	dirLimit := 4
	if len(dirs) < dirLimit {
		dirLimit = len(dirs)
	}
	if m.fileCount < dirLimit {
		dirLimit = m.fileCount
	}
	for i := 0; i < dirLimit; i++ {
		icon := "  "
		name := dirs[i]
		if len(name) > 28 {
			name = name[:25] + "..."
		}
		if i == dirLimit-1 && !m.scanDone {
			content += themeGlow(t).Render(icon+name) + "\n"
		} else {
			content += themeFolder(t).Render(icon) + " " + themeFile(t).Render(name) + "\n"
		}
	}
	if len(dirs) > dirLimit {
		content += lipgloss.NewStyle().Foreground(t.muted).Render(fmt.Sprintf("    ... and %d more", len(dirs)-dirLimit)) + "\n"
	}

	content += "\n" + themeTitle(t).Render(" Files ") + " " + strings.Repeat("─", 33) + "\n"

	fileLimit := m.fileCount - dirLimit
	if fileLimit < 0 {
		fileLimit = 0
	}
	if fileLimit > len(files) {
		fileLimit = len(files)
	}

	cols := 2
	colWidth := 22
	var rows []string
	row := ""
	col := 0

	showCount := fileLimit

	for i := 0; i < showCount; i++ {
		name := files[i]
		display := " " + name
		if len(display) > colWidth-2 {
			display = display[:colWidth-2]
		}

		if i == showCount-1 && !m.scanDone {
			display = themeGlow(t).Render(display)
		}

		row += display
		col++

		if col >= cols || i == showCount-1 {
			rows = append(rows, row)
			row = ""
			col = 0
		} else {
			row += "  "
		}
	}

	for _, r := range rows {
		content += r + "\n"
	}

	if len(files) > showCount {
		content += lipgloss.NewStyle().Foreground(t.muted).Render(fmt.Sprintf("    ... and %d more files", len(files)-showCount)) + "\n"
	}

	if !m.scanDone {
		content += "\n  " + m.spinner.View() + " loading..."
	}

	stats := fmt.Sprintf("%d folders  •  %d files", len(dirs), len(files))
	content += "\n\n" + lipgloss.NewStyle().Foreground(t.primary).Render(strings.Repeat("─", w-8)) + "\n"
	content += "  " + lipgloss.NewStyle().Foreground(t.muted).Render(stats)

	return lipgloss.NewStyle().Width(w).Height(18).Padding(0, 0, 0, 1).
		Border(lipgloss.ThickBorder()).BorderForeground(t.primary).
		Render(content)
}

func (m *model) renderMonitor(t Theme) string {
	w := 88
	var s string

	s += themeTitle(t).Render(" System Information ") + "\n\n"

	s += fmt.Sprintf("  %-12s %s\n", "Hostname:", themeGlow(t).Render(m.hostName))
	s += fmt.Sprintf("  %-12s %s\n", "Kernel:", themeGlow(t).Render(m.kernelVer))
	s += fmt.Sprintf("  %-12s %s\n", "OS:", themeGlow(t).Render(m.osInfo))
	s += fmt.Sprintf("  %-12s %s\n", "CPUs:", themeGlow(t).Render(fmt.Sprintf("%d", m.numCPUs)))

	gpuStr := "None detected"
	if len(m.gpus) > 0 {
		var gpuNames []string
		for _, g := range m.gpus {
			entry := g.name
			if g.curFreqMHz > 0 || g.maxFreqMHz > 0 {
				entry += fmt.Sprintf(" (%d/%d MHz)", g.curFreqMHz, g.maxFreqMHz)
			}
			gpuNames = append(gpuNames, entry)
		}
		gpuStr = strings.Join(gpuNames, ", ")
	}
	s += fmt.Sprintf("  %-12s %s\n", "GPU:", themeGlow(t).Render(gpuStr))

	s += fmt.Sprintf("  %-12s %s\n", "Uptime:", themeGlow(t).Render(formatUptime(m.uptime)))
	s += fmt.Sprintf("  %-12s %.2f / %.2f / %.2f\n", "Load Avg:", m.loadAvg[0], m.loadAvg[1], m.loadAvg[2])

	s += "\n" + themeTitle(t).Render(" Resources ") + "\n\n"

	cpuBar := m.renderBar(m.cpuPercent, 36, t)
	memBar := m.renderBar(m.memPercent, 36, t)
	swapBar := m.renderBar(m.swapPercent, 36, t)
	diskBar := m.renderBar(m.diskPct, 36, t)

	memLabel := fmt.Sprintf("%d / %d MB", m.memUsed, m.memTotal)
	swapLabel := "No Swap"
	if m.swapTotal > 0 {
		swapLabel = fmt.Sprintf("%d / %d MB", m.swapUsed, m.swapTotal)
	}
	diskLabel := fmt.Sprintf("%d / %d GB", m.diskUsed, m.diskTotal)

	s += fmt.Sprintf("  CPU:   %s  %5.1f%%\n", cpuBar, m.cpuPercent)
	s += fmt.Sprintf("  RAM:   %s  %5.1f%%  %s\n", memBar, m.memPercent, m.renderSideInfo(memLabel, m.memPercent, t))
	s += fmt.Sprintf("  Swap:  %s  %s\n", swapBar, m.renderSideInfo(swapLabel, m.swapPercent, t))
	s += fmt.Sprintf("  Disk:  %s  %5.1f%%  %s\n", diskBar, m.diskPct, m.renderSideInfo(diskLabel, m.diskPct, t))

	s += "\n" + themeTitle(t).Render(" Top Processes (by RAM) ") + "\n\n"

	s += fmt.Sprintf("  %-7s %-18s %6s  %8s  %5s\n", "PID", "NAME", "STATE", "MEM MB", "MEM%")
	s += lipgloss.NewStyle().Foreground(t.muted).Render("  ─────────────────────────────────────────────────") + "\n"

	for _, p := range m.procs {
		name := p.command
		if len(name) > 16 {
			name = name[:13] + "..."
		}
		state := procState(p.state, t)
		s += fmt.Sprintf("  %-7d %-18s %s  %7.1f MB  %4.1f%%\n", p.pid, name, state, p.memMB, p.memPct)
	}

	s += "\n" + lipgloss.NewStyle().Foreground(t.muted).Render(" ──────────────────────────────────────────────────────────────")
	s += "\n  " + lipgloss.NewStyle().Foreground(t.muted).Italic(true).Render("Press Tab to switch panels")

	return lipgloss.NewStyle().Width(w).Padding(0, 2).
		Render(s)
}

func (m *model) renderTheme(t Theme) string {
	w := 88
	var s string

	s += themeTitle(t).Render(" Color Themes ") + "\n\n"
	s += lipgloss.NewStyle().Foreground(t.muted).Render("  Use ↑/↓ arrows to browse themes\n")
	s += lipgloss.NewStyle().Foreground(t.muted).Render("  Press Tab to go back\n\n")

	start := 0
	if m.themeIndex >= 5 {
		start = m.themeIndex - 4
	}
	end := start + 9
	if end > len(themeList) {
		end = len(themeList)
		start = end - 9
		if start < 0 {
			start = 0
		}
	}

	for i := start; i < end; i++ {
		theme := themes[themeList[i]]
		idx := i + 1
		ptr := "  "
		mark := "  "

		if i == m.themeIndex {
			ptr = "▶ "
		}

		if m.theme == themeList[i] {
			mark = "● "
		}

		preview := lipgloss.NewStyle().
			Foreground(theme.primary).Render("██")
		preview += " " + lipgloss.NewStyle().Foreground(theme.secondary).Render("██")
		preview += " " + lipgloss.NewStyle().Foreground(theme.success).Render("██")
		preview += " " + lipgloss.NewStyle().Foreground(theme.warning).Render("██")

		line := fmt.Sprintf("%s%s %s %s %s", ptr, mark, preview, lipgloss.NewStyle().Foreground(t.dimText).Render(fmt.Sprintf("%2d.", idx)), theme.name)
		s += line + "\n"
	}

	s += "\n" + lipgloss.NewStyle().Foreground(t.muted).Render(" ──────────────────────────────────────────────────────────────")

	curTheme := applyTheme(m.theme)
	s += "\n  " + lipgloss.NewStyle().Foreground(t.muted).Render("Current: ") + lipgloss.NewStyle().Foreground(curTheme.primary).Bold(true).Render(curTheme.name)

	s += "\n\n  " + lipgloss.NewStyle().Foreground(t.muted).Italic(true).Render("Themes: Ocean, Forest, Sunset, Aurora, Lavender,")
	s += "\n  " + lipgloss.NewStyle().Foreground(t.muted).Italic(true).Render("Coral, Dusk, Slate, Rose, Amber, Indigo, Teal,")
	s += "\n  " + lipgloss.NewStyle().Foreground(t.muted).Italic(true).Render("Cherry, Sage, Midnight")

	return lipgloss.NewStyle().Width(w).Padding(0, 2).
		Render(s)
}

func procState(s string, t Theme) string {
	switch s {
	case "R":
		return themeBarLow(t).Render("R")
	case "S":
		return themeGlow(t).Render("S")
	case "D":
		return lipgloss.NewStyle().Foreground(t.warning).Render("D")
	case "Z":
		return lipgloss.NewStyle().Foreground(t.error).Render("Z")
	case "T":
		return lipgloss.NewStyle().Foreground(t.secondary).Render("T")
	default:
		return s
	}
}

func (m *model) renderBar(pct float64, width int, t Theme) string {
	if pct > 100 {
		pct = 100
	}
	filled := int(float64(width) * pct / 100)
	if filled > width {
		filled = width
	}
	if pct > 0 && filled == 0 {
		filled = 1
	}

	var b strings.Builder
	b.WriteString("[")
	for i := 0; i < width; i++ {
		if i < filled {
			if pct > 85 {
				b.WriteString("█")
			} else if pct > 60 {
				b.WriteString("█")
			} else {
				b.WriteString("█")
			}
		} else {
			b.WriteString(" ")
		}
	}
	b.WriteString("]")

	var col lipgloss.Color
	if pct > 85 {
		col = t.error
	} else if pct > 60 {
		col = t.warning
	} else {
		col = t.success
	}

	return lipgloss.NewStyle().Foreground(col).Render(b.String())
}

func (m *model) renderSideInfo(text string, pct float64, t Theme) string {
	var col lipgloss.Color
	if pct > 85 {
		col = t.error
	} else if pct > 60 {
		col = t.warning
	} else {
		col = t.muted
	}
	return lipgloss.NewStyle().Foreground(col).Render(text)
}

func (m *model) safeClean(t *target) error {
	if t.path == "/nix/store" {
		return nil
	}
	if strings.HasPrefix(t.path, "/nix") {
		return nil
	}

	entries, err := os.ReadDir(t.path)
	if err != nil {
		return err
	}
	for _, e := range entries {
		fp := filepath.Join(t.path, e.Name())

		info, err := os.Lstat(fp)
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}

		rel, err := filepath.Rel(t.path, fp)
		if err != nil || strings.HasPrefix(rel, "../") {
			continue
		}

		_ = os.RemoveAll(fp)
	}
	return nil
}

func hasRootPrivileges() bool {
	return os.Geteuid() == 0
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func getLuminaDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func (m *model) update() tea.Cmd {
	return func() tea.Msg {
		dir := getLuminaDir()

		pullCmd := exec.Command("git", "pull")
		pullCmd.Dir = dir
		pullOut, err := pullCmd.CombinedOutput()
		if err != nil {
			return updateMsg{success: false, output: fmt.Sprintf("git pull failed:\n%s", string(pullOut))}
		}

		buildCmd := exec.Command("go", "build", "-o", "lumina", ".")
		buildCmd.Dir = dir
		buildOut, err := buildCmd.CombinedOutput()
		if err != nil {
			return updateMsg{success: false, output: fmt.Sprintf("build failed:\n%s", string(buildOut))}
		}

		return updateMsg{success: true, output: fmt.Sprintf("git pull:\n%s\nbuild: success", string(pullOut))}
	}
}

func (m *model) clean() tea.Cmd {
	return func() tea.Msg {
		var errs []string
		selectedAny := false

		for _, t := range m.targets {
			if t.selected && t.safe {
				selectedAny = true
				if err := m.safeClean(t); err != nil {
					errs = append(errs, fmt.Sprintf("%s: %v", t.label, err))
				}
			}
		}

		if !selectedAny {
			return cleanDoneMsg{errors: errs}
		}

		osID := strings.ToLower(getOSID())

		var cmds [][]string
		switch osID {
		case "ubuntu", "debian", "linuxmint", "pop", "zorin", "elementary", "deepin", "raspbian":
			cmds = [][]string{{"apt-get", "clean"}, {"apt-get", "autoremove", "-y"}}
		case "fedora", "centos", "rocky", "almalinux":
			cmds = [][]string{{"dnf", "clean", "all"}}
		case "arch", "archarm", "manjaro", "endeavouros", "cachyos":
			if commandExists("paccache") {
				cmds = [][]string{{"paccache", "-r", "-k0"}}
			} else {
				cmds = [][]string{{"pacman", "-Sc", "--noconfirm"}}
			}
		case "void":
			cmds = [][]string{{"xbps-remove", "-O"}}
		case "alpine":
			cmds = [][]string{{"apk", "cache", "clean"}}
		case "nixos":
			cmds = [][]string{{"nix-collect-garbage", "-d"}}
		case "opensuse", "opensuse-leap", "opensuse-tumbleweed":
			cmds = [][]string{{"zypper", "clean", "-a"}}
		case "gentoo":
			cmds = [][]string{{"eclean-distfiles"}}
		default:
			entries, err := os.ReadDir("/tmp")
			if err == nil {
				for _, e := range entries {
					_ = os.RemoveAll(filepath.Join("/tmp", e.Name()))
				}
			}
		}

		for _, args := range cmds {
			cmd := exec.Command(args[0], args[1:]...)
			if out, err := cmd.CombinedOutput(); err != nil {
				msg := strings.TrimSpace(string(out))
				if msg != "" {
					errs = append(errs, fmt.Sprintf("%s: %s", args[0], msg))
				} else {
					errs = append(errs, fmt.Sprintf("%s: %v", args[0], err))
				}
			}
		}

		return cleanDoneMsg{errors: errs}
	}
}

type tickMsg struct{}
type countdownMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*60, func(t time.Time) tea.Msg { return tickMsg{} })
}

func startCountdown() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return countdownMsg{} })
}

func animTick() tea.Cmd {
	return tea.Tick(time.Millisecond*25, func(t time.Time) tea.Msg { return animTickMsg{} })
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--update" {
		dir := getLuminaDir()
		fmt.Println("Updating Lumina...")

		pullCmd := exec.Command("git", "pull")
		pullCmd.Dir = dir
		pullOut, err := pullCmd.CombinedOutput()
		fmt.Print(string(pullOut))
		if err != nil {
			fmt.Printf("git pull failed: %v\n", err)
			os.Exit(1)
		}

		buildCmd := exec.Command("go", "build", "-o", "lumina", ".")
		buildCmd.Dir = dir
		buildOut, err := buildCmd.CombinedOutput()
		fmt.Print(string(buildOut))
		if err != nil {
			fmt.Printf("build failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Lumina updated successfully!")
		return
	}

	if !hasRootPrivileges() {
		fmt.Println("This program requires root privileges. Please run with sudo or pkexec.")
		os.Exit(1)
	}

	osName, osIcon := detectOS()
	osID := getOSID()
	targets := buildTargets(osID)

	t, i, nCPUs := getCPUUsage()

	cfg := loadConfig()
	themeIdx := 0
	for i, thm := range themeList {
		if thm == cfg.Theme {
			themeIdx = i
			break
		}
	}

	m := &model{
		spinner:      spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(lipgloss.NewStyle().Foreground(themes[cfg.Theme].primary))),
		progress:     progress.New(progress.WithGradient(string(themes[cfg.Theme].primary), string(themes[cfg.Theme].secondary))),
		targets:      targets,
		state:        "ready",
		osInfo:       osName,
		osIcon:       osIcon,
		sizing:       true,
		prevCPUTotal: t,
		prevCPUIdle:  i,
		numCPUs:      nCPUs,
		loadAvg:      []float64{0, 0, 0},
		theme:        cfg.Theme,
		themeIndex:   themeIdx,
	}

	getSysInfo(m)

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
