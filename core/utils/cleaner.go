package utils

import (
	"log"
)

func cleanIllumina(seqs []Sequence, details bool) ([]Sequence, error) {
	adapters, err := loadAdapters("config/adapters.json")
	qualities, errQ := loadQualities("config/quality.json")
	if err != nil {
		log.Fatalf("Error to load adapters: %v", err)
	}
	if errQ != nil {
		log.Fatalf("Error to load qualities: %v", err)
	}
	var cleaned []Sequence
	for _, seq := range seqs {
		seq = trimAdapters(seq, adapters["Illumina"])
		if validateQuality(seq.Quality, qualities["Illumina"].Threshold, qualities["Illumina"].MaxBadBases) && isValidLength(seq.Bases, qualities["Illumina"].Minbases) && !hasHomopolymer(seq.Bases, qualities["Illumina"].Homo) {
			cleaned = append(cleaned, seq)
		}
	}
	return cleaned, err
}

func cleanNanopore(seqs []Sequence, details bool) ([]Sequence, error) {
	adapters, err := loadAdapters("config/adapters.json")
	qualities, errQ := loadQualities("config/quality.json")
	if err != nil {
		log.Fatalf("Error to load adapters: %v", err)
	}
	if errQ != nil {
		log.Fatalf("Error to load qualities: %v", err)
	}
	var cleaned []Sequence
	for _, seq := range seqs {
		seq = trimAdapters(seq, adapters["OxfordNanopore"])
		if validateQuality(seq.Quality, qualities["OxfordNanopore"].Threshold, qualities["OxfordNanopore"].MaxBadBases) && isValidLength(seq.Bases, qualities["OxfordNanopore"].Minbases) && !hasHomopolymer(seq.Bases, qualities["OxfordNanopore"].Homo) {
			cleaned = append(cleaned, seq)
		}
	}
	return cleaned, err
}

func cleanPacBio(seqs []Sequence, details bool) ([]Sequence, error) {
	adapters, err := loadAdapters("config/adapters.json")
	qualities, errQ := loadQualities("config/quality.json")
	if err != nil {
		log.Fatalf("Error to load adapters: %v", err)
	}
	if errQ != nil {
		log.Fatalf("Error to load qualities: %v", err)
	}
	var cleaned []Sequence
	for _, seq := range seqs {
		seq = trimAdapters(seq, adapters["PacBio"])
		if validateQuality(seq.Quality, qualities["PacBio"].Threshold, qualities["PacBio"].MaxBadBases) && isValidLength(seq.Bases, qualities["PacBio"].Minbases) && !hasHomopolymer(seq.Bases, qualities["PacBio"].Homo) {
			cleaned = append(cleaned, seq)
		}
	}
	return cleaned, err
}

func cleanIonTorrent(seqs []Sequence, details bool) ([]Sequence, error) {
	adapters, err := loadAdapters("config/adapters.json")
	qualities, errQ := loadQualities("config/quality.json")
	if err != nil {
		log.Fatalf("Error to load adapters: %v", err)
	}
	if errQ != nil {
		log.Fatalf("Error to load qualities: %v", err)
	}
	var cleaned []Sequence
	for _, seq := range seqs {
		seq = trimAdapters(seq, adapters["IonTorrent"])
		if validateQuality(seq.Quality, qualities["IonTorrent"].Threshold, qualities["IonTorrent"].MaxBadBases) && isValidLength(seq.Bases, qualities["IonTorrent"].Minbases) && !hasHomopolymer(seq.Bases, qualities["IonTorrent"].Homo) {
			cleaned = append(cleaned, seq)
		}
	}
	return cleaned, err
}
