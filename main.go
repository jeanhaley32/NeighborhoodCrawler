package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Jeanhaley32/neighborfinder"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

var (
	jsonpath, writefile string
	runs                = 0
)

func init() {
	flag.StringVar(&jsonpath, "jsonpath", "node-list.json", "Path to json file containing bootnodes. Defauklt is ./node-list.json")
	flag.StringVar(&writefile, "writefile", "finalized-node-list.json", "Path to write json file containing nodes with added neighbors. Default is ./finalized-node-list.json")
	flag.Parse()
}

// nodeMap is a map of node IDs to ENR records.
type nodeMap map[string]enrJSON

// JSON struct representing an ENR record
type enrJSON struct {
	Seq           uint64        `json:"seq"`
	Record        string        `json:"record"`
	Score         int           `json:"score"`
	FirstResponse string        `json:"firstResponse"`
	LastResponse  string        `json:"lastResponse"`
	LastCheck     string        `json:"lastCheck"`
	Neighbors     []*enode.Node `json:"neighbors"`
}

func main() {
	// Ensure jsonpath and writefile are not empty.
	switch {
	case jsonpath == "":
		log.Fatalf("json path is empty")
	case writefile == "":
		log.Fatalf("write path is empty")
	}
	// Open the JSON Node file.
	file, err := os.Open(jsonpath)
	if err != nil {
		log.Fatalf("Failed to open JSON file: %s", err.Error())
	}

	// Decode the JSON file into a list of ENR records.'
	var entries nodeMap
	err = json.NewDecoder(file).Decode(&entries) // decode the JSON file into the entries map
	if err != nil {
		log.Fatalf("Failed to decode JSON file: %s", err.Error())
	}

	// Iterate through the entries map, and add neighbors to each entry.
	// Then write the entries map to the writefile.
	for _, entry := range entries {
		entry.Neighbors = neighborfinder.Getneighbors(entry.Record)
		runs++
		nodeTrunc := fmt.Sprintf("%v...%v", entry.Record[5:10], entry.Record[len(entry.Record)-5:])
		clearScreen()
		fmt.Printf("ID: %v \nFound %v neighbors\nrun %v/%v\n", nodeTrunc, len(entry.Neighbors), runs, len(entries))
	}

	// write the entries map to the writefile.
	writeJsonToFile(entries, writefile)
}

// writeJsonToFile writes a JSON file to the writefile.
func writeJsonToFile(d any, outputfile string) {
	file, err := os.Create(outputfile)
	if err != nil {
		log.Fatalf("Failed to create file: %s", err.Error())
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(d)
	if err != nil {
		log.Fatalf("Failed to encode JSON file: %s", err.Error())
	}
}

// exit prints the error to stderr and exits with status 1.
func exit(err interface{}) {
	if err == nil {
		os.Exit(0)
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// decodeRecordHex decodes a hex-encoded node record.
func decodeRecordHex(b []byte) ([]byte, bool) {
	if bytes.HasPrefix(b, []byte("0x")) {
		b = b[2:]
	}
	dec := make([]byte, hex.DecodedLen(len(b)))
	_, err := hex.Decode(dec, b)
	return dec, err == nil
}

// Clears Terminal Screen.
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
