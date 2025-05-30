package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Result struct {
	Cache string
	Chunk int
	Time  float64
	RAM   int
	CPU   float64
}

func main() {
	input := "./benchmark/sequences_examples/SRR33681711.fastq"
	chunks := []int{50000, 100000, 200000}
	caches := []string{"ram", "disk"}

	outputDir := "./benchmark"
	os.MkdirAll(outputDir+"/results", 0o755)

	fmt.Printf("=== Start benchmark MARIA (Go) ===\n")
	fmt.Printf("🧠 Available cores: %d\n\n", runtime.NumCPU())

	var results []Result

	for _, chunk := range chunks {
		for _, cache := range caches {
			diskFlag := "false"
			if cache == "disk" {
				diskFlag = "true"
			}

			outfile := fmt.Sprintf("%s/results/out_%s_%d.fastq", outputDir, cache, chunk)
			logfile := fmt.Sprintf("%s/results/log_%s_%d.txt", outputDir, cache, chunk)

			cmd := exec.Command("./maria",
				"-in", input,
				"-out", outfile,
				"-disk="+diskFlag,
				"-chunk="+strconv.Itoa(chunk),
			)

			log, err := os.Create(logfile)
			if err != nil {
				fmt.Println("Error to create file log:", err)
				continue
			}
			cmd.Stderr = log
			start := time.Now()
			fmt.Printf("▶️ Execution: cache=%s | chunk=%d\n", cache, chunk)

			if err := cmd.Run(); err != nil {
				fmt.Println("❌ Error execution:", err)
				continue
			}
			duration := time.Since(start)
			log.Close()

			// read last log
			file, _ := os.Open(logfile)
			scanner := bufio.NewScanner(file)
			var last string
			for scanner.Scan() {
				last = scanner.Text()
			}
			file.Close()

			parts := strings.Split(last, ";")
			if len(parts) != 3 {
				fmt.Println("❌ Format incompatible:", last)
				continue
			}

			ram, _ := strconv.Atoi(parts[1])
			cpuStr := strings.TrimSpace(strings.TrimSuffix(parts[2], "%"))
			cpu, _ := strconv.ParseFloat(cpuStr, 64)

			results = append(results, Result{
				Cache: cache, Chunk: chunk, Time: duration.Seconds(), RAM: ram, CPU: cpu,
			})

			fmt.Printf("✅ Time: %.2fs | RAM: %d KB | CPU: %.1f%%\n\n", duration.Seconds(), ram, cpu)
		}
	}

	// Save results
	out, _ := os.Create(outputDir + "/summary.csv")
	defer out.Close()
	out.WriteString("Cache,ChunkSize,Time(s),RAM(KB),CPU(%)\n")
	for _, r := range results {
		out.WriteString(fmt.Sprintf("%s,%d,%.2f,%d,%.1f\n", r.Cache, r.Chunk, r.Time, r.RAM, r.CPU))
	}

	fmt.Println("🏁 Benchmark complete. Results on benchmarks/summary.csv")
	// execute
	// go run benchmark/main.go
}
