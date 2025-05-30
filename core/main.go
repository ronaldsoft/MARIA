package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"MARIA/core/utils"
)

func main() {
	fmt.Println("Hello to MARIA: A novel bioinformatic toolkit making on golang to clean sequences")
	input := flag.String("in", "", "(.fastq, .fq, .fasta, .fa) -> File compatible with: Illumina, Oxford Nanopore, PacBio, and Ion Torrent")
	output := flag.String("out", "", "Path of clean file")
	pluginList := flag.String("plugins", "", "List of plugins separate for comma (order acendent execution)")
	preWorker := flag.Bool("preworker", false, "Active order per worker execution")
	details := flag.Bool("details", false, "Generates reports and folders with files of: adapters, invalid sequences, primers and low-quality sequences.")
	threads := flag.Int("threads", 0, "Number of threads for use (0 use all)")
	useDisk := flag.Bool("disk", false, "Use disk cache (default RAM)")
	chunkSize := flag.Int("chunk", 0, "Number of lines per chunk")
	flag.Parse()

	if *input == "" || *output == "" {
		fmt.Println("Use with Go: go run core/main.go -in raw(.fastq, .fq, .fasta, .fa) -out clean.fastq -plugins=compressFile -disk=true -chunk=100000")
		fmt.Println("Use with build: ./maria -in raw(.fastq, .fq, .fasta, .fa) -out clean.fastq -plugins=compressFile -disk=true -chunk=100000")
		os.Exit(1)
	}
	fileFormat, fileLines := utils.CheckFileFormat(*input)
	sample, err := utils.PeekFirstReads(*input, 100)
	if err != nil {
		log.Fatalf("Error read file")
		return
	}
	// check tecnology
	fmt.Printf("Cores Aveables: %d\n", utils.AvailableCPU())
	fmt.Printf("RAM total: %d GB\n", utils.AvailableRAM()/1e9)
	fmt.Printf("RAM usable (~75%%): %d GB\n", utils.UsableRAM()/1e9)

	if *chunkSize == 0 {
		size, totalChunks, memory := utils.AutoEstimateChunks(*input, fileLines)
		fmt.Printf("Lines per chunk: %d (%.2f MB per core)\n", size, memory)
		fmt.Printf("Total chunks: %d\n", totalChunks)
		*chunkSize = size
	}
	tech := utils.DetectSequencingTech(sample)
	if tech == "" {
		log.Fatalf("Error secuence technology")
	}
	fmt.Println("Technology detect:", tech)
	utils.NextPhase("Valid type technology", 1)
	ramOK := utils.SystemHasEnoughRAM()
	nvme := utils.IsNVMeMounted()
	useDiskCache := !ramOK && nvme
	fmt.Printf("RAM: %v | NVMe: %v | Cache on disk: %v\n", ramOK, nvme, useDiskCache)

	utils.NextPhase("Generating temporal directory", 2)
	tempDir := filepath.Join(os.TempDir(), "maria_clean_chunks")
	fmt.Printf("Temporal files on: %v \n", tempDir)
	os.MkdirAll(tempDir, 0o755)

	utils.NextPhase("Valid format of secuence", 3)
	if fileFormat == "fastq" {
		utils.ParallelClean(*input, *output, *chunkSize, tech, *useDisk, *threads, tempDir, *pluginList, *preWorker, *details)
	} else if fileFormat == "fasta" {
		utils.ParallelClean(*input, *output, *chunkSize, tech, *useDisk, *threads, tempDir, *pluginList, *preWorker, *details)
	} else {
		fmt.Println("Format not supported")
	}

	fmt.Println("Thank for Use MARIA - Finish process.")
}
