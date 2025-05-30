package utils

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

func ParallelClean(
	inputPath,
	outputPath string,
	chunkSize int,
	tech string,
	useDisk bool,
	threads int,
	tempDir string,
	pluginList string,
	preWorker bool,
	details bool,
) {
	if threads <= 0 {
		threads = AvailableCPU()
	}
	fmt.Printf("Threads: %v\n", threads)
	reader, err := SmartReadFile(inputPath)
	if err != nil {
		log.Fatalf("Error open file: %v", err)
	}
	jobs := make(chan [][4]string, threads*2)
	var wg sync.WaitGroup
	NextPhase("Run on parallel threads", 4)
	// Launches workers
	startWorkers(threads, jobs, tempDir, tech, &wg, pluginList, preWorker, details)
	// process all chunks generates
	processChunks(reader, jobs, chunkSize)
	wg.Wait()
	// generate file output
	handleOutput(outputPath, tempDir, pluginList)
}

func DetectSequencingTech(lines []string) string {
	for _, l := range lines {
		line := strings.ToLower(l)
		// Illumina: cabezal típico con @NS o patrón típico de Illumina
		if strings.Contains(l, "@NS") || strings.Contains(l, ":1:") {
			return "Illumina"
			// Oxford Nanopore: típico contiene "runid" en encabezados
		} else if strings.Contains(l, "@") && strings.Contains(l, "runid") {
			return "Oxford Nanopore"
			// PacBio: líneas que inician con ">m" o contienen "ccs"
		} else if strings.HasPrefix(l, ">m") || strings.Contains(l, "ccs") {
			return "PacBio"
			// Ion Torrent: patron más específico, por ejemplo "S1" en id, pero cuidado
			// Aquí puedes agregar más patrones específicos según tu data
		} else if strings.Contains(l, "S1") ||
			strings.Contains(l, "R_") ||
			strings.Contains(line, "ionxpress") ||
			strings.Contains(line, "ion") ||
			strings.Contains(line, "thermo") ||
			strings.Contains(line, "s1") || // usualmente lane o sample
			strings.Contains(line, "pgm") || // modelo Ion Torrent Personal Genome Machine
			strings.Contains(line, "s5") {
			return "Ion Torrent"
		}
	}
	return ""
}

func mergeChunks(tempDir, outputPath string) error {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error to create file: %w", err)
	}
	defer outputFile.Close()

	files, err := filepath.Glob(filepath.Join(tempDir, "chunk_*.tmp"))
	if err != nil {
		return fmt.Errorf("error to list tmp files: %w", err)
	}

	sort.Strings(files) // sort files to merge in order

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("error read %s: %w", file, err)
		}
		if _, err := outputFile.Write(data); err != nil {
			return fmt.Errorf("error write ouput file: %w", err)
		} else {
			DeleteTempFile(file)
		}
	}

	return nil
}

func cleanRead(read [4]string, tech string, details bool) [4]string {
	seq := Sequence{
		ID:      strings.TrimSpace(read[0]),
		Bases:   strings.TrimSpace(read[1]),
		Plus:    strings.TrimSpace(read[2]),
		Quality: strings.TrimSpace(read[3]),
	}

	var cleaned []Sequence
	switch tech {
	case "Illumina":
		cleaned, _ = cleanIllumina([]Sequence{seq}, details)
	case "Oxford Nanopore":
		cleaned, _ = cleanNanopore([]Sequence{seq}, details)
	case "PacBio":
		cleaned, _ = cleanPacBio([]Sequence{seq}, details)
	case "Ion Torrent":
		cleaned, _ = cleanIonTorrent([]Sequence{seq}, details)
	default:
		log.Fatalf("not found tech")
		cleaned = []Sequence{seq} // sin limpieza si no se reconoce
	}

	if len(cleaned) > 0 {
		c := cleaned[0]
		return [4]string{
			c.ID + "\n",
			c.Bases + "\n",
			c.Plus + "\n",
			c.Quality + "\n",
		}
	}

	// No pass, return empty
	return [4]string{"", "", "", ""}
}

func CheckFileFormat(filename string) (string, int) {
	fileFormat := ""
	fileLines := 2
	format := strings.ToLower(filepath.Ext(filename))
	if format == ".fastq" || format == ".fq" {
		fileFormat = "fastq"
		fileLines = 4
	} else if format == ".fasta" || format == ".fa" {
		fileFormat = "fasta"
		fileLines = 2
	} else {
		fileFormat = "fasta"
		fileLines = 2
	}
	return fileFormat, fileLines
}

func startWorkers(threads int, jobs <-chan [][4]string, tempDir, tech string, wg *sync.WaitGroup, pluginList string, preWorker bool, details bool) {
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(id int) {
			fmt.Printf("Worker %d started\n", id)
			defer wg.Done()
			for chunk := range jobs {
				fmt.Printf("Worker %d received chunk with %d sequences\n", id, len(chunk))
				for idx, read := range chunk {
					cleaned := cleanRead(read, tech, details)
					// execute prev actions to each read
					if preWorker {
						cleaned = ExecuteToWorkersPlugins(pluginList, cleaned)
					}
					if cleaned[0] == "" {
						continue
					}
					WriteTempFile(tempDir, fmt.Sprintf("chunk_%d_seq_%d.tmp", id, idx), strings.Join(cleaned[:], ""))
				}
			}
		}(i)
	}
}

func processChunks(reader *bufio.Reader, jobs chan<- [][4]string, chunkSize int) {
	var chunk [][4]string

	for {
		var seq [4]string
		readLines := 0
		for i := 0; i < 4; i++ {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				if readLines > 0 {
					chunk = append(chunk, seq)
				}
				if len(chunk) > 0 {
					fmt.Printf("Sent last chunk of size %d to jobs\n", len(chunk))
					jobs <- chunk
				}
				close(jobs)
				return
			}
			if err != nil {
				log.Println("Error reading line:", err)
				continue
			}
			seq[i] = line
			readLines++
		}
		if readLines == 4 {
			chunk = append(chunk, seq)
		}
		if len(chunk) >= chunkSize {
			fmt.Printf("Sent chunk of size %d to jobs\n", len(chunk))
			jobs <- chunk
			chunk = nil
		}
	}
}

func handleOutput(outputPath, tempDir, pluginList string) {
	NextPhase("Generating file output", 5)
	err := mergeChunks(tempDir, outputPath)
	if err != nil {
		log.Fatalf("Error merging chunks: %v", err)
	}
	fmt.Printf("Files are merged: %v\n", outputPath)

	fmt.Println("Clean sequences complete")
	if pluginList != "" {
		NextPhase("Start to run plugins", 6)
		ExecutePlugins(pluginList, outputPath)
	}
	fmt.Println("All phases completed.")
}
