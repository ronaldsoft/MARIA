package utils

import "log"

func cleanIllumina(seqs []Sequence) ([]Sequence, error) {
	adapters, err := loadAdapters("config/adapters.json")
	if err != nil {
		log.Fatalf("Error to load adapters: %v", err)
	}
	var cleaned []Sequence
	for _, seq := range seqs {
		seq = trimAdapters(seq, adapters["Illumina"])
		if meanQuality(seq.Quality) >= 25 && len(seq.Bases) >= 50 {
			cleaned = append(cleaned, seq)
		}
	}
	return cleaned, err
}

func cleanNanopore(seqs []Sequence) ([]Sequence, error) {
	adapters, err := loadAdapters("config/adapters.json")
	if err != nil {
		log.Fatalf("Error to load adapters: %v", err)
	}
	var cleaned []Sequence
	for _, seq := range seqs {
		seq = trimAdapters(seq, adapters["OxfordNanopore"])
		if meanQuality(seq.Quality) >= 12 && len(seq.Bases) >= 500 && !hasHomopolymer(seq.Bases, 10) {
			cleaned = append(cleaned, seq)
		}
	}
	return cleaned, err
}

func cleanPacBio(seqs []Sequence) ([]Sequence, error) {
	adapters, err := loadAdapters("config/adapters.json")
	if err != nil {
		log.Fatalf("Error to load adapters: %v", err)
	}
	var cleaned []Sequence
	for _, seq := range seqs {
		seq = trimAdapters(seq, adapters["PacBio"])
		if meanQuality(seq.Quality) >= 20 && len(seq.Bases) >= 500 {
			cleaned = append(cleaned, seq)
		}
	}
	return cleaned, err
}

func cleanIonTorrent(seqs []Sequence) ([]Sequence, error) {
	adapters, err := loadAdapters("config/adapters.json")
	if err != nil {
		log.Fatalf("Error to load adapters: %v", err)
	}
	var cleaned []Sequence
	for _, seq := range seqs {
		seq = trimAdapters(seq, adapters["IonTorrent"])
		if meanQuality(seq.Quality) >= 20 && len(seq.Bases) >= 100 && !hasHomopolymer(seq.Bases, 8) {
			cleaned = append(cleaned, seq)
		}
	}
	return cleaned, err
}
