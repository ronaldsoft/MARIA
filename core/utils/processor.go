package utils

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

func ParallelClean(inputPath, outputPath string, chunkSize int, tech string, useDisk bool, threads int, tempDir string, pluginList string) {
	file, err := os.Open(inputPath)
	if err != nil {
		log.Fatalf("Error open file: %v", err)
	}
	defer file.Close()

	if threads <= 0 {
		threads = runtime.NumCPU()
	}

	fmt.Printf("Threads: %v\n", threads)

	reader := bufio.NewReader(file)
	jobs := make(chan [][4]string, threads*2)
	var wg sync.WaitGroup

	NextPhase("Run on parallel threads", 4)

	// Launches workers
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(id int) {
			fmt.Printf("Worker %d started\n", id)
			defer wg.Done()
			for chunk := range jobs {
				fmt.Printf("Worker %d receive chunk with %d secuences\n", id, len(chunk))
				for idx, read := range chunk {
					cleaned := cleanRead(read, tech)
					fmt.Printf("Worker %d clean secuence %d \n", id, idx)

					if cleaned[0] == "" {
						continue // skip empty
					}
					fmt.Printf("Worker %d writing on disk folder tmp\n", id)
					WriteTempFile(tempDir, fmt.Sprintf("chunk_%d_seq_%d.tmp", id, idx), strings.Join(cleaned[:], ""))
				}
			}
		}(i)
	}

	// read file
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
					fmt.Printf("Sent last chunk of size %d to channel jobs\n", len(chunk))
					jobs <- chunk
				}
				fmt.Println("Close channel jobs")
				close(jobs)
				wg.Wait()

				NextPhase("Generating file output", 5)
				err := mergeChunks(tempDir, outputPath)
				if err != nil {
					log.Fatalf("Error merge chunks: %v\n", err)
				} else {
					fmt.Printf("Files are merged: %v\n", outputPath)
				}

				fmt.Println("Clean sequences complete")
				if pluginList != "" {
					NextPhase("Start to run plugins", 6)
					ExecutePlugins(pluginList, outputPath)
				}
				fmt.Println("All phases completed.")
				return
			}
			if err != nil {
				log.Println("Error read line:", err)
				continue
			}
			seq[i] = line
			readLines++
		}
		if readLines == 4 {
			chunk = append(chunk, seq)
		}
		if len(chunk) >= chunkSize {
			fmt.Printf("sent chunk of size %d of channel jobs\n", len(chunk))
			jobs <- chunk
			chunk = nil
		}
	}
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

func cleanRead(read [4]string, tech string) [4]string {
	seq := Sequence{
		ID:      strings.TrimSpace(read[0]),
		Bases:   strings.TrimSpace(read[1]),
		Plus:    strings.TrimSpace(read[2]),
		Quality: strings.TrimSpace(read[3]),
	}

	var cleaned []Sequence
	switch tech {
	case "Illumina":
		cleaned, _ = cleanIllumina([]Sequence{seq})
	case "Oxford Nanopore":
		cleaned, _ = cleanNanopore([]Sequence{seq})
	case "PacBio":
		cleaned, _ = cleanPacBio([]Sequence{seq})
	case "Ion Torrent":
		cleaned, _ = cleanIonTorrent([]Sequence{seq})
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
