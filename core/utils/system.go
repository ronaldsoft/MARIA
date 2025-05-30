package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"plugin"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/mem"
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
		fmt.Printf("Executing plugin: %s\n", pluginPath)

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

func ExecuteToWorkersPlugins(pluginList string, cleanedSecuences [4]string) [4]string {
	plugins := strings.Split(pluginList, ",")
	seqs := cleanedSecuences
	for _, pluginName := range plugins {
		pluginPath := fmt.Sprintf("plugins/%s.so", pluginName)
		fmt.Printf("Executing to workers plugin: %s\n", pluginPath)

		p, err := plugin.Open(pluginPath)
		if err != nil {
			log.Printf("Can't open plugin %s: %v\n", pluginName, err)
			continue
		}

		sym, err := p.Lookup("ProcessDataWorker")
		if err != nil {
			log.Printf("Plugin %s nt have function 'ProcessDataWorker'\n", pluginName)
			continue
		}

		processFunc, ok := sym.(func([4]string) ([4]string, error))
		if !ok {
			log.Printf("'ProcessDataWorker' on plugin %s have a incorrect type\n", pluginName)
			continue
		}

		if seqs, err = processFunc(cleanedSecuences); err != nil {
			log.Printf("Plugin %s fail: %v\n", pluginName, err)
			continue
		}
	}
	return seqs
}

func NextPhase(title string, phaseCounter int) {
	fmt.Printf("\nPhase %d: %s...\n", phaseCounter, title)
	phaseCounter++
}

func AvailableCPU() int {
	cpus := runtime.NumCPU()
	return cpus
}

func UsableRAM() uint64 {
	return uint64(float64(AvailableRAM()) * 0.75) // 25% free
}

func AvailableRAM() uint64 {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("Error obtain memort info:", err)
		return 0
	}
	return uint64(vmem.Total)
}

func availableMemPerCore() float64 {
	memPerCore := UsableRAM() / uint64(AvailableCPU())
	return float64(memPerCore)
}

func targetMemPerThread() float64 {
	totalRAMMB := AvailableRAM()
	cores := AvailableCPU()
	var targetMemPerThread float64
	switch {
	// chunk 500 – 1,000 reads
	case totalRAMMB <= 4096 && cores <= 4:
		targetMemPerThread = 300 * 1024
		// chunk 1,000 – 5,000 reads
	case totalRAMMB <= 8192 && cores <= 4:
		targetMemPerThread = 1024 * 1024
		// 5,000 – 10,000+ reads
	case totalRAMMB > 8192 && cores >= 4:
		targetMemPerThread = 4 * 1024 * 1024
	default:
		targetMemPerThread = 512 * 1024
	}
	return targetMemPerThread
}

func countLinesAndAvgSize(filename string, linesPerSeq int) (int, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 1000, 0, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	totalLines := 0
	totalBytes := 0
	chunkSize := 0
	for scanner.Scan() {
		line := scanner.Text()
		totalBytes += len(line) + 1 // +1 by '\n
		totalLines++
	}
	if err := scanner.Err(); err != nil {
		return 1000, 0, err
	}
	if totalLines == 0 {
		return 1000, 0, nil
	}
	// max chunks
	maxChunks := AvailableCPU() * 10
	minChunkSize := totalLines / maxChunks
	if minChunkSize > chunkSize {
		chunkSize = minChunkSize
	}
	// min chunk control
	if chunkSize < 10 {
		chunkSize = 10
	}
	avgLineSize := float64(totalBytes) / float64(totalLines)

	avgSeqSize := avgLineSize * float64(linesPerSeq)
	chunkSize = int(targetMemPerThread() / avgSeqSize)
	totalChunks := int(math.Ceil(float64(totalLines) / float64(chunkSize)))
	return chunkSize, totalChunks, nil
}

func AutoEstimateChunks(filepath string, lines int) (int, int, float64) {
	// read file
	chunkSize, totalChunks, err := countLinesAndAvgSize(filepath, lines)
	if err != nil {
		fmt.Println("Error read file:", err)
		return 0, 0, 0
	}
	return chunkSize, totalChunks, float64(availableMemPerCore()) / 1024.0 / 1024.0
}

func SmartReadFile(filepath string) (*bufio.Reader, error) {
	maxSize := UsableRAM()

	info, err := infoFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error getting file info: %w", err)
	}

	var reader *bufio.Reader
	// dynamic limit based on RAM Avaiable
	if info.Size() <= int64(maxSize) {
		data, err := os.ReadFile(filepath)
		if err != nil {
			return nil, fmt.Errorf("error reading file: %w", err)
		}
		reader = bufio.NewReader(bytes.NewReader(data))
	} else {
		// read on buffer big size file
		file, err := os.Open(filepath)
		if err != nil {
			return nil, fmt.Errorf("error opening file: %w", err)
		}
		file.Close() // need close
		reader = bufio.NewReader(file)
	}

	return reader, nil
}

func infoFile(filepath string) (os.FileInfo, error) {
	info, err := os.Stat(filepath)
	if err != nil {
		fmt.Println("Can't access file:", err)
		return nil, nil
	}
	return info, err
}
