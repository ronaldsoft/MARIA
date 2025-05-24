package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"plugin"
	"runtime"
	"strings"
)

func PeekFirstReads(path string, n int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() && len(lines) < n*4 {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}

func IsNVMeMounted() bool {
	switch runtime.GOOS {
	case "linux":
		return LinuxIsNVMeMounted()
	// case "windows":
	// 	return checkWindowsRAM()
	case "darwin":
		return DarwinNVMeMounted()
	default:
		return false
	}
}

func LinuxIsNVMeMounted() bool {
	f, _ := os.Open("/proc/mounts")
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		if strings.Contains(s.Text(), "/dev/nvme") {
			return true
		}
	}
	return false
}

func DarwinNVMeMounted() bool {
	out, err := exec.Command("diskutil", "list").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "NVMe")
}

func WindowsNVMeMounted() bool {
	return false
}

func SystemHasEnoughRAM() bool {
	switch runtime.GOOS {
	case "linux":
		return checkLinuxRAM()
	// case "windows":
	// 	return checkWindowsRAM()
	case "darwin":
		return checkDarwinRAM()
	default:
		return false
	}
}

// Linux
func checkLinuxRAM() bool {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return false
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				var kb int
				fmt.Sscanf(fields[1], "%d", &kb)
				return kb > 4000000 // 4 GB
			}
		}
	}
	return false
}

// // Windows
// func checkWindowsRAM() bool {
// 	// Utiliza syscall para obtener memoria total
// 	var memStatus struct {
// 		Length               uint32
// 		MemoryLoad           uint32
// 		TotalPhys            uint64
// 		AvailPhys            uint64
// 		TotalPageFile        uint64
// 		AvailPageFile        uint64
// 		TotalVirtual         uint64
// 		AvailVirtual         uint64
// 		AvailExtendedVirtual uint64
// 	}
// 	memStatus.Length = uint32(unsafe.Sizeof(memStatus))
// 	ret, _, _ := syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx").Call(uintptr(unsafe.Pointer(&memStatus)))
// 	if ret == 0 {
// 		return false
// 	}
// 	return memStatus.TotalPhys > 4*1024*1024*1024 // 4 GB
// }

// macOS
func checkDarwinRAM() bool {
	// Utiliza sysctl para obtener memoria física
	const hwMemsize = "hw.memsize"
	out, err := exec.Command("sysctl", "-n", hwMemsize).Output()
	if err != nil {
		return false
	}
	var bytes int64
	fmt.Sscanf(string(out), "%d", &bytes)
	return bytes > 4*1024*1024*1024 // 4 GB
}

func ExecutePlugins(pluginList string, cleanedFilePath string) {
	plugins := strings.Split(pluginList, ",")
	for _, pluginName := range plugins {
		pluginPath := fmt.Sprintf("plugins/%s.so", pluginName)
		fmt.Printf("🔧 Executing plugin: %s\n", pluginPath)

		p, err := plugin.Open(pluginPath)
		if err != nil {
			log.Printf("Can't open plugin %s: %v\n", pluginName, err)
			continue
		}

		sym, err := p.Lookup("Process")
		if err != nil {
			log.Printf("Plugin %s nt have function 'Process'\n", pluginName)
			continue
		}

		processFunc, ok := sym.(func(string) error)
		if !ok {
			log.Printf("'Process' on plugin %s have a incorrect type\n", pluginName)
			continue
		}

		if err := processFunc(cleanedFilePath); err != nil {
			log.Printf("Plugin %s fail: %v\n", pluginName, err)
		}
	}
}

func NextPhase(title string, phaseCounter int) {
	fmt.Printf("\nPhase %d: %s...\n", phaseCounter, title)
	phaseCounter++
}
