package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type Result struct {
	Cache string
	Chunk int
	Time  float64
	RAM   int
	CPU   float64
}

func check_benchmarks() {
	input := "/mnt/data/reads_1M.fastq"
	format := "fastq"
	chunks := []int{50000, 100000, 200000}
	caches := []string{"ram", "disk"}

	outputDir := "benchmarks"
	os.MkdirAll(outputDir+"/results", 0o755)

	fmt.Printf("=== Iniciando benchmark NGS Cleaner (Go) ===\n")
	fmt.Printf("🧠 Núcleos disponibles: %d\n\n", runtime.NumCPU())

	var results []Result

	for _, chunk := range chunks {
		for _, cache := range caches {
			diskFlag := "false"
			if cache == "disk" {
				diskFlag = "true"
			}

			outfile := fmt.Sprintf("%s/results/out_%s_%d.fastq", outputDir, cache, chunk)
			logfile := fmt.Sprintf("%s/results/log_%s_%d.txt", outputDir, cache, chunk)

			cmd := exec.Command("/usr/bin/time", "-f", "%e;%M;%P",
				"./ngscleaner",
				"-in", input,
				"-out", outfile,
				"-fmt", format,
				"-disk="+diskFlag,
				"-chunk="+strconv.Itoa(chunk),
			)

			log, err := os.Create(logfile)
			if err != nil {
				fmt.Println("Error creando archivo de log:", err)
				continue
			}
			cmd.Stderr = log

			fmt.Printf("▶️ Ejecutando: cache=%s | chunk=%d\n", cache, chunk)
			// start := time.Now()
			if err := cmd.Run(); err != nil {
				fmt.Println("❌ Error en ejecución:", err)
				continue
			}
			log.Close()

			// Leer último log
			file, _ := os.Open(logfile)
			scanner := bufio.NewScanner(file)
			var last string
			for scanner.Scan() {
				last = scanner.Text()
			}
			file.Close()

			parts := strings.Split(last, ";")
			if len(parts) != 3 {
				fmt.Println("❌ Formato inesperado:", last)
				continue
			}

			etime, _ := strconv.ParseFloat(parts[0], 64)
			ram, _ := strconv.Atoi(parts[1])
			cpuStr := strings.TrimSpace(strings.TrimSuffix(parts[2], "%"))
			cpu, _ := strconv.ParseFloat(cpuStr, 64)

			results = append(results, Result{
				Cache: cache, Chunk: chunk, Time: etime, RAM: ram, CPU: cpu,
			})

			fmt.Printf("✅ Tiempo: %.2fs | RAM: %d KB | CPU: %.1f%%\n\n", etime, ram, cpu)
		}
	}

	// Guardar resultados
	out, _ := os.Create(outputDir + "/summary.csv")
	defer out.Close()
	out.WriteString("Cache,ChunkSize,Time(s),RAM(KB),CPU(%)\n")
	for _, r := range results {
		out.WriteString(fmt.Sprintf("%s,%d,%.2f,%d,%.1f\n", r.Cache, r.Chunk, r.Time, r.RAM, r.CPU))
	}

	fmt.Println("🏁 Benchmark completado. Resultados en benchmarks/summary.csv")
}
