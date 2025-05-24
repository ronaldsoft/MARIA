package utils

import (
	"encoding/json"
	"os"
	"strings"
)

type Sequence struct {
	ID      string
	Bases   string
	Plus    string
	Quality string
}

// Recorta cualquier adaptador encontrado en la lista de `adapters`.
// Utiliza strings.Index (usa Boyer-Moore internamente).
func trimAdapters(seq Sequence, adapters []string) Sequence {
	for _, adapter := range adapters {
		pos := strings.Index(seq.Bases, adapter)
		if pos != -1 {
			seq.Bases = seq.Bases[:pos]
			if len(seq.Quality) > pos {
				seq.Quality = seq.Quality[:pos]
			} else {
				seq.Quality = seq.Quality // por si acaso
			}
			break // cortar al primer adaptador encontrado
		}
	}
	return seq
}

func meanQuality(q string) float64 {
	var sum int
	for _, c := range q {
		sum += int(c) - 33 // phred+33
	}
	return float64(sum) / float64(len(q))
}

func hasHomopolymer(seq string, maxLen int) bool {
	count := 1
	for i := 1; i < len(seq); i++ {
		if seq[i] == seq[i-1] {
			count++
			if count > maxLen {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}

func loadAdapters(filename string) (map[string][]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var adapters map[string][]string
	err = json.Unmarshal(data, &adapters)
	if err != nil {
		return nil, err
	}
	return adapters, nil
}
