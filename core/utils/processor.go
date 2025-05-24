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
	// results := make(chan [4]string, chunkSize*threads)
	var wg sync.WaitGroup

	NextPhase("Run on parallel threads", 4)

	// Lanzar workers
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(id int) {
			fmt.Printf("Worker %d iniciado\n", id)
			defer wg.Done()
			for chunk := range jobs {
				fmt.Printf("Worker %d recibió chunk con %d secuencias\n", id, len(chunk))
				for idx, read := range chunk {
					cleaned := cleanRead(read, tech)
					fmt.Printf("Worker %d limpia secuencia %d \n", id, idx)

					if cleaned[0] == "" {
						continue // skip empty
					}
					// if useDisk {
					fmt.Printf("Worker %d escribe en disco tmp\n", id)
					WriteTempFile(tempDir, fmt.Sprintf("chunk_%d_seq_%d.tmp", id, idx), strings.Join(cleaned[:], ""))
					// } else {
					// 	fmt.Printf("Worker %d usa memoria\n", id)
					// 	results <- cleaned
					// }
				}
			}
		}(i)
	}

	// Lector de archivo
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
					fmt.Printf("Enviando último chunk de tamaño %d al canal jobs\n", len(chunk))
					jobs <- chunk
				}
				fmt.Println("Cerrando canal jobs")
				close(jobs)
				wg.Wait()

				// Escribir resultados
				// if useDisk {
				NextPhase("Generating file output", 5)
				err := mergeChunks(tempDir, outputPath)
				if err != nil {
					log.Fatalf("Error merge chunks: %v", err)
				} else {
					fmt.Println("Files are merged cleaned_output.fastq")
				}
				// } else {
				// 	outputFile, err := os.Create(outputPath)
				// 	if err != nil {
				// 		log.Fatalf("Error create output file: %v", err)
				// 	}
				// 	defer outputFile.Close()

				// 	go func() {
				// 		wg.Wait()
				// 		close(results)
				// 	}()

				// 	for read := range results {
				// 		_, err := outputFile.WriteString(read[0] + read[1] + read[2] + read[3])
				// 		if err != nil {
				// 			log.Fatalf("Error write file: %v", err)
				// 		}
				// 	}
				// 	NextPhase("Generating file output", 5)
				// 	fmt.Printf("File %v generated in memory.\n", outputPath)
				// }

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
			fmt.Printf("Enviando chunk de tamaño %d al canal jobs\n", len(chunk))
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
