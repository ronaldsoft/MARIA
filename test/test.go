package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Parámetros de limpieza
const (
	minQualityChar = '!' + 10 // Phred+33, mínimo calidad = 10
	minLength      = 50       // longitud mínima permitida
	adapter        = "TGGCAG" // ejemplo de adaptador (opcional)
)

// Limpia una secuencia Oxford Nanopore
func cleanOxfordRead(read [4]string) (cleaned [4]string) {
	header := strings.TrimSpace(read[0])
	seq := strings.TrimSpace(read[1])
	plus := strings.TrimSpace(read[2])
	qual := strings.TrimSpace(read[3])

	// Validar largo mínimo
	if len(seq) < minLength || len(qual) != len(seq) {
		return
	}

	// 1. Eliminar adaptador si está presente
	if strings.Contains(seq, adapter) {
		parts := strings.Split(seq, adapter)
		seq = parts[0]
		qual = qual[:len(seq)] // cortar calidad también
	}

	// 2. Validar bases válidas (A,C,G,T,N)
	for _, base := range seq {
		if !strings.ContainsRune("ACGTacgt", base) {
			return
		}
	}

	// 3. Recortar extremos 3' de baja calidad
	trimmedSeq := ""
	trimmedQual := ""
	for i := 0; i < len(seq); i++ {
		if qual[i] < byte(minQualityChar) {
			break
		}
		trimmedSeq += string(seq[i])
		trimmedQual += string(qual[i])
	}

	if len(trimmedSeq) < minLength {
		return
	}

	// 4. Setear read limpio
	cleaned[0] = header + "\n"
	cleaned[1] = trimmedSeq + "\n"
	cleaned[2] = plus + "\n"
	cleaned[3] = trimmedQual + "\n"
	return
}

// Procesa un archivo FASTQ
func processFastq(input, output string) error {
	in, err := os.Open(input)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(output)
	if err != nil {
		return err
	}
	defer out.Close()

	reader := bufio.NewReader(in)
	writer := bufio.NewWriter(out)
	defer writer.Flush()

	for {
		var read [4]string
		for i := 0; i < 4; i++ {
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil // fin del archivo
			}
			read[i] = line
		}

		cleaned := cleanOxfordRead(read)
		if cleaned[0] != "" {
			for _, line := range cleaned {
				writer.WriteString(line)
			}
		}
	}
}

func main() {
	input := "sample_1_ontarget_nanopore.fastq"
	output := "cleaned_output.fastq"

	err := processFastq(input, output)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Archivo limpiado generado:", output)
}
