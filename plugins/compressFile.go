package main

import (
	"fmt"
	"os"
	"os/exec"
)

// Exported function required by the system
func Process(path string) error {
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
