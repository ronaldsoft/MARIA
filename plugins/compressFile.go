package main

import (
	"fmt"
	"os"
	"os/exec"
)

// Exported function required by the system
func Process(path string) error {
	// this lib only compatible with linux and mac
	zipPath := path + ".gz"
	cmd := exec.Command("gzip", "-c", path)
	outFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("no se pudo crear el archivo comprimido: %v", err)
	}
	defer outFile.Close()

	cmd.Stdout = outFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("falló la compresión gzip: %v", err)
	}
	fmt.Println("✅ Archivo comprimido en:", zipPath)
	return nil
}

// structure for use fasta
// data[0]: ID = Identifier
// data[1]: Bases = Nucleotides

// structure for use fastq
// data[0]: ID = Identifier
// data[1]: Bases = Nucleotides
// data[2]: Plus = Extra
// data[3]: Quality = Quality Nucleotide on secuence
func ProcessDataWorker(data [4]string) ([4]string, error) {
	fmt.Printf("my secuence id: %v", data[0])
	return data, nil
}
