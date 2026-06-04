package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
)

type triageFile struct {
	Accepted []triageEntry `json:"accepted"`
}

type triageEntry struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Issue    string `json:"issue"`
	Owner    string `json:"owner"`
	Decision string `json:"decision"`
}

type govulnEvent struct {
	Config  *json.RawMessage `json:"config"`
	Finding *struct {
		OSV string `json:"osv"`
	} `json:"finding"`
}

func main() {
	triagePath := flag.String("triage", "security/govulncheck-triage.json", "path to govulncheck triage JSON")
	inputPath := flag.String("input", "-", "path to govulncheck JSON stream, or - for stdin")
	flag.Parse()

	accepted, err := loadTriage(*triagePath)
	if err != nil {
		fail(err)
	}

	input, closeInput, err := openInput(*inputPath)
	if err != nil {
		fail(err)
	}
	defer closeInput()

	found, sawConfig, err := readFindings(input)
	if err != nil {
		fail(err)
	}
	if !sawConfig {
		fail(fmt.Errorf("govulncheck JSON stream did not include a config event"))
	}

	var untriaged []string
	for id := range found {
		if _, ok := accepted[id]; !ok {
			untriaged = append(untriaged, id)
		}
	}
	sort.Strings(untriaged)
	if len(untriaged) > 0 {
		fail(fmt.Errorf("untriaged govulncheck findings: %v", untriaged))
	}

	ids := make([]string, 0, len(found))
	for id := range found {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	fmt.Printf("govulncheck triage passed; findings=%v\n", ids)
}

func openInput(path string) (io.Reader, func(), error) {
	if path == "-" {
		return os.Stdin, func() {}, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	return f, func() { _ = f.Close() }, nil
}

func loadTriage(path string) (map[string]triageEntry, error) {
	bz, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var triage triageFile
	if err := json.Unmarshal(bz, &triage); err != nil {
		return nil, err
	}
	accepted := make(map[string]triageEntry, len(triage.Accepted))
	for _, entry := range triage.Accepted {
		if entry.ID == "" || entry.Severity == "" || entry.Issue == "" || entry.Owner == "" || entry.Decision == "" {
			return nil, fmt.Errorf("triage entry %q must include id, severity, issue, owner, and decision", entry.ID)
		}
		if _, exists := accepted[entry.ID]; exists {
			return nil, fmt.Errorf("duplicate triage entry for %s", entry.ID)
		}
		accepted[entry.ID] = entry
	}
	return accepted, nil
}

func readFindings(r io.Reader) (map[string]struct{}, bool, error) {
	dec := json.NewDecoder(r)
	found := map[string]struct{}{}
	sawConfig := false
	for {
		var event govulnEvent
		if err := dec.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			return nil, false, err
		}
		if event.Config != nil {
			sawConfig = true
		}
		if event.Finding != nil && event.Finding.OSV != "" {
			found[event.Finding.OSV] = struct{}{}
		}
	}
	return found, sawConfig, nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
