package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const defaultSource = "https://raw.githubusercontent.com/cncf/landscape/master/landscape.yml"

func main() {
	out := flag.String("out", "landscape.yml", "output path")
	source := flag.String("source", defaultSource, "official CNCF Landscape YAML URL")
	flag.Parse()

	client := &http.Client{Timeout: 90 * time.Second}
	response, err := client.Get(*source)
	if err != nil {
		fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		fatal(fmt.Errorf("download %s: %s", *source, response.Status))
	}
	payload, err := io.ReadAll(response.Body)
	if err != nil {
		fatal(err)
	}
	if len(payload) < 100 || string(payload[:3]) == "404" {
		fatal(fmt.Errorf("downloaded CNCF Landscape payload is unexpectedly small or invalid"))
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*out, payload, 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("Updated %s from %s (%d bytes)\n", *out, *source, len(payload))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
