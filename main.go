package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gosom/google-maps-scraper/scraper"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var (
		inputFile   string
		outputFile  string
		concurrency int
		depth       int
		lang        string
		geoCoords   string
		zoom        int
		webServer   bool
		webAddr     string
		printVersion bool
	)

	flag.StringVar(&inputFile, "input", "", "path to input file with search queries (one per line)")
	flag.StringVar(&outputFile, "output", "", "path to output CSV file (default: stdout)")
	// bumped default concurrency from 1 to 3 for faster scraping on my machine
	flag.IntVar(&concurrency, "concurrency", 3, "number of concurrent scrapers")
	flag.IntVar(&depth, "depth", 10, "max depth of results per query")
	flag.StringVar(&lang, "lang", "en", "language code for Google Maps (e.g. en, de, fr)")
	flag.StringVar(&geoCoords, "geo", "", "geo coordinates for search bias (e.g. '37.7749,-122.4194')")
	flag.IntVar(&zoom, "zoom", 15, "zoom level for map (1-21)")
	flag.BoolVar(&webServer, "web", false, "start web server for job management")
	flag.StringVar(&webAddr, "web-addr", ":8080", "address for web server to listen on")
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")
	flag.Parse()

	if printVersion {
		fmt.Printf("google-maps-scraper version=%s commit=%s date=%s\n", version, commit, date)
		os.Exit(0)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "received shutdown signal, stopping...")
		cancel()
	}()

	cfg := scraper.Config{
		InputFile:   inputFile,
		OutputFile:  outputFile,
		Concurrency: concurrency,
		Depth:       depth,
		Lang:        lang,
		GeoCoords:   geoCoords,
		Zoom:        zoom,
		WebServer:   webServer,
		WebAddr:     webAddr,
	}

	if err := run(ctx, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg scraper.Config) error {
	app, err := scraper.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize scraper: %w", err)
	}
	defer app.Close()

	return app.Run(ctx)
}
