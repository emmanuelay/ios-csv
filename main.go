package main

import (
	"flag"
	"log"
	"os"
)

type Config struct {
	InPath    string
	OutPath   string
	Delimiter string
	Verbose   bool
	ToCSV     bool
}

func main() {
	config := parseFlags()

	// Set up logging
	if config.Verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(0)
	}

	log.Println("Processing ...")

	processor := NewLocalizationProcessor(config)
	if config.ToCSV {
		if err := processor.ConvertToCSV(); err != nil {
			log.Printf("Error converting to CSV: %v", err)
			os.Exit(1)
		}
	} else {
		if err := processor.ConvertToIOS(); err != nil {
			log.Printf("Error converting to iOS localized strings: %v", err)
			os.Exit(1)
		}
	}
}

func parseFlags() Config {
	var config Config
	flag.StringVar(&config.InPath, "i", "", "Input source path (required)")
	flag.StringVar(&config.OutPath, "o", "", "Output path (required)")
	flag.StringVar(&config.Delimiter, "d", ",", "Delimiter used as separation of values")
	flag.BoolVar(&config.Verbose, "v", false, "Increase output verbosity")
	flag.BoolVar(&config.ToCSV, "tocsv", false, "Convert from iOS localized folders to CSV (default is CSV to iOS)")
	flag.Parse()

	if config.InPath == "" || config.OutPath == "" {
		flag.Usage()
		os.Exit(1)
	}
	return config
}
