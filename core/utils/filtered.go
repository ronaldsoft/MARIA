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

type QualityThresholds map[string]struct {
	Threshold   int `json:"threshold"`
	Minbases    int `json:"minbases"`
	Homo        int `json:"homopolymer"`
	MaxBadBases int `json:"maxBadBases"`
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
			}
			break // cortar al primer adaptador encontrado
		}
	}
	return seq
}

// valid quality of nucleotids with fastq
// detectPhredOffset detects the Phred quality score offset (33 or 64) from the quality string.
func detectPhredOffset(qual string) int {
	// Count characters above 73 (higher ASCII)
	count64 := 0
	for i := 0; i < len(qual); i++ {
		if qual[i] > 73 {
			count64++
		}
	}
	if count64 > len(qual)/2 {
		return 64
	}
	return 33
}

// decodePhred converts a single ASCII character to its Phred score using the detected offset.
func decodePhred(qualChar byte, offset int) int {
	return int(qualChar) - offset
}

// validateQuality returns true if all quality scores are equal or above the threshold.
func validateQuality(quality string, threshold int, maxBadBases int) bool {
	offset := detectPhredOffset(quality)
	// include tolerance
	badCount := 0
	for i := 0; i < len(quality); i++ {
		score := decodePhred(quality[i], offset)
		if score < threshold {
			badCount++
		}
	}
	// valid tolerance
	if badCount > maxBadBases {
		return false // too many bad bases
	}
	return true
}

func isValidLength(seq string, minLen int) bool {
	return len(seq) >= minLen
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

func loadQualities(filename string) (QualityThresholds, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var thresholds QualityThresholds
	err = json.Unmarshal(data, &thresholds)
	return thresholds, err
}
