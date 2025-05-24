# MARIA: A novel bioinformatics toolkit making on Golang to clean sequences

**MARIA** An ultrafast, multithreaded bioinformatics toolkit developed in Go, designed for efficient preprocessing and cleaning of raw sequencing data from **Illumina**, **Oxford Nanopore**, **PacBio**, and **Ion Torrent** technologies.

**Why MARIA?**

- My heart: _"This tool bears the name of the one who gave me half of my life, twenty-three chromosomes of profound love, and the eternal breath of her glowing mitochondria. Thanks to you, **Katlenth María Esdrina Carranza Montesdeoca**, I am what I am, because your essence vibrates in every cell of me"._
- Anthropology: María is one of the most common names in the world.
- Symbolism: Purity, maternal love, humility, and faith.

---

## 🚀 Features

- ⚡ Blazing fast multithreaded processing
- 🧬 Supports FASTQ and other raw sequence formats
- 🧹 Quality filtering, trimming, and adapter removal
- 🧪 Compatible with short and long reads
- 🔧 CLI-based, lightweight, and cross-platform
- 🔄 Designed for integration in pipelines

---

## 📦 Installation

### Requirements

- Go 1.21+
- Git

### Clone and build

```bash
git clone https://github.com/ronaldsoft/MARIA.git
cd MARIA
go build -o maria ./core/main.go
```

## 🛠 Usage

```bash
./maria -in sample_1_ontarget_nanopore.fastq -out secuenciasCleaned.fastq -plugins=compressFile
```

### Recommendations Based on RAM and Number of Cores

This document describes the optimal `chunkSize` for cleaning DNA/RNA sequences (FASTQ or FASTA format) on systems with limited resources.

| Available RAM | CPU Cores | Recommended Chunk Size (reads) | Estimated Memory per Thread |
| ------------- | --------- | ------------------------------ | --------------------------- |
| ≤ 4 GB        | 2–4       | 500 – 1,000                    | ~200–400 KB                 |
| 4 – 8 GB      | 2–4       | 1,000 – 5,000                  | ~400 KB – 2 MB              |
| > 8 GB        | ≥ 4       | 5,000 – 10,000+                | > 2 MB                      |

### Notes

- Each FASTQ read contains 4 lines (identifier, bases, separator, quality scores).
- Estimated memory usage: 1 read ≈ 400 bytes.
- Using larger chunks may improve performance on machines with ample RAM, but may cause bottlenecks or swapping on limited systems.
- It is recommended to start testing with `chunkSize = 1000` and adjust based on system behavior.

## 🛠 Generate plugins

```bash
go build -buildmode=plugin -o plugins/compressFile.so plugins/compressFile.go
go run main.go -plugins=compressFile,customPlugin
```

## Contributors

- [Ronald Rivera](https://github.com/ronaldsoft)

## Contact

[Create an issue](https://github.com/ronaldsoft/MARIA/issues) to report bugs,
propose new functions or ask for help.

<!-- ## License

[MIT License](https://github.com/ronaldsoft/MARIA/blob/master/LICENSE) -->

## Benchmark comparative

## Citation
