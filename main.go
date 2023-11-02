package main

import (
	"encoding/json"
	"flag"
	"fmt"
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
	Seq           uint64       `json:"seq"`
	Record        string       `json:"record"`
	Score         int          `json:"score"`
	FirstResponse string       `json:"firstResponse"`
	LastResponse  string       `json:"lastResponse"`
	LastCheck     string       `json:"lastCheck"`
	Neighbors     []enode.Node `json:"neighbors"`
}

func main() {
	// Ensure jsonpath and writefile are not empty.
	switch {
	case jsonpath == "":
		exit(fmt.Errorf("json path is empty"))
	case writefile == "":
		exit(fmt.Errorf("write path is empty"))
	}
	// Open the JSON Node file.
	file, err := os.Open(jsonpath)
	if err != nil {
		exit(fmt.Errorf("Failed to open JSON file: %s", err.Error()))
	}

	// Decode the JSON file into a list of ENR records.'
	var entries nodeMap
	err = json.NewDecoder(file).Decode(&entries) // decode the JSON file into the entries map
	if err != nil {
		exit(fmt.Errorf("Failed to decode JSON file: %s", err.Error()))
	}

	// Iterate through the entries map, and add neighbors to each entry.
	// Then write the entries map to the writefile.
	for id, entry := range entries {
		pointerbucket := neighborfinder.Getneighbors(entry.Record)
		neighbors := make([]enode.Node, len(pointerbucket))
		for i, pointer := range pointerbucket {
			neighbors[i] = *pointer
		}
		entry.Neighbors = append(entry.Neighbors, neighbors...)
		entries[id] = entry
		runs++
		nodeTrunc := fmt.Sprintf("%v...%v", entry.Record[5:10], entry.Record[len(entry.Record)-5:])
		clearScreen()
		fmt.Printf("ID: %v \nFound %v neighbors\nrun %v/%v\n", nodeTrunc, len(entry.Neighbors), runs, len(entries))
		fmt.Println("------------------------------------------------------------------")
		for _, neighbor := range entry.Neighbors {
			fmt.Printf("Neighbor: %v...%v\n", neighbor.ID().String()[5:10], neighbor.ID().String()[len(neighbor.ID().String())-5:])
		}
	}

	clearScreen()
	// // write the entries map to the writefile.
	err = writeJsonToFile(entries, writefile)
	if err != nil {
		exit(err)
	}

}

// writeJsonToFile writes a JSON file to the writefile.
func writeJsonToFile(d any, outputfile string) error {
	file, err := os.Create(outputfile)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %s", err.Error())
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(d)
	if err != nil {
		return fmt.Errorf("failed to encode JSON file: %s", err.Error())
	}
	return nil
}

// exit prints the error to stderr and exits with status 1.
func exit(err interface{}) {
	if err == nil {
		os.Exit(0)
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// Clears Terminal Screen.
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
