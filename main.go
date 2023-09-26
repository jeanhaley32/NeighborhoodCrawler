package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	jsonpath, writefile string
	enrRecordPrefix     = []byte("enr:-")
	runs                = 0
)

const (
	cs = "\033[H\033[2J"
)

func init() {
	flag.StringVar(&jsonpath, "jsonpath", "node-list.json", "Path to json file containing bootnodes. Defauklt is ./node-list.json")
	flag.StringVar(&writefile, "writefile", "finalized-node-list.json", "Path to write json file containing nodes with added neighbors. Default is ./finalized-node-list.json")
	flag.Parse()
}

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
	// Check if the jsonpath and writefile are empty.
	// If they are empty, exit with a fatal error.
	if jsonpath == "" {
		log.Fatalf("jsonpath is empty")
	}
	if writefile == "" {
		log.Fatalf("writepath is empty")
	}

	// // Create a nodetree to store the nodes and their neighbors.
	// var nt nodetree

	// Open the JSON file containing the bootnodes.
	file, err := os.Open(jsonpath)
	if err != nil {
		log.Fatalf("Failed to open JSON file: %s", err.Error())
	}

	// Decode the JSON file into a list of ENR records.'
	var entries nodeMap                          // map of string to enrJSON struct for each entry
	err = json.NewDecoder(file).Decode(&entries) // decode the JSON file into the entries map
	if err != nil {
		log.Fatalf("Failed to decode JSON file: %s", err.Error())
	}

	// populate target with the enode nodes from the JSON file.
	for _, entry := range entries {
		neighbors := getNeighbors(entry.Record)
		entry.Neighbors = neighbors
		runs++
		// fmt.Print(clearScreen())
		fmt.Printf("ID: %v \n Runs: %d\n Found %v neighbors\n", entry.Record, runs, len(neighbors))
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

// // startV4 starts an ephemeral discovery V4 node.
func startV4(nodekey, bootnodes, nodedb, extaddr string) (*discover.UDPv4, discover.Config, error) {
	ln, config := makeDiscoveryConfig(nodekey, nodedb)
	socket := listen(ln, extaddr)
	disc, err := discover.ListenV4(socket, ln, config)
	if err != nil {
		return nil, config, err
	}
	return disc, config, nil
}

// makeDiscoveryConfig creates a discovery configuration.
// A discovery configuration is used to create a discovery node.
func makeDiscoveryConfig(nodekey, nodedb string) (*enode.LocalNode, discover.Config) {
	var cfg discover.Config

	if nodekey != "" {
		key, err := crypto.HexToECDSA(nodekey)
		if err != nil {
			exit(fmt.Errorf("-%s: %v", nodekey, err))
		}
		cfg.PrivateKey = key
	} else {
		cfg.PrivateKey, _ = crypto.GenerateKey()
	}

	dbpath := nodedb
	db, err := enode.OpenDB(dbpath)
	if err != nil {
		exit(err)
	}
	ln := enode.NewLocalNode(db, cfg.PrivateKey)
	return ln, cfg
}

func listen(ln *enode.LocalNode, extAddr string) *net.UDPConn {
	addr := "0.0.0.0:0"
	socket, err := net.ListenPacket("udp4", addr)
	if err != nil {
		exit(err)
	}

	// Configure UDP endpoint in ENR from listener address.
	usocket := socket.(*net.UDPConn)
	uaddr := socket.LocalAddr().(*net.UDPAddr)
	if uaddr.IP.IsUnspecified() {
		ln.SetFallbackIP(net.IP{127, 0, 0, 1})
	} else {
		ln.SetFallbackIP(uaddr.IP)
	}
	ln.SetFallbackUDP(uaddr.Port)

	if extAddr != "" {
		ip, port, ok := parseExtAddr(extAddr)
		if !ok {
			exit(fmt.Errorf("invalid external address %q", extAddr))
		}
		ln.SetStaticIP(ip)
		if port != 0 {
			ln.SetFallbackUDP(port)
		}
	}

	return usocket
}

// exit prints the error to stderr and exits with status 1.
func exit(err interface{}) {
	if err == nil {
		os.Exit(0)
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// parseExtAddr parses an external address specification.
func parseExtAddr(spec string) (ip net.IP, port int, ok bool) {
	ip = net.ParseIP(spec)
	if ip != nil {
		return ip, 0, true
	}
	host, portstr, err := net.SplitHostPort(spec)
	if err != nil {
		return nil, 0, false
	}
	ip = net.ParseIP(host)
	if ip == nil {
		return nil, 0, false
	}
	port, err = strconv.Atoi(portstr)
	if err != nil {
		return nil, 0, false
	}
	return ip, port, true
}

// parseBootnodes parses a comma-separated list of bootnodes.
func parseBootnodes(bootNodes string) ([]*enode.Node, error) {
	s := params.MainnetBootnodes
	if bootNodes != "" {
		input := bootNodes
		if input == "" {
			return nil, nil
		}
		s = strings.Split(input, ",")
	}
	nodes := make([]*enode.Node, len(s))
	var err error
	for i, record := range s {
		nodes[i], err = parseNode(record)
		if err != nil {
			return nil, fmt.Errorf("invalid bootstrap node: %v", err)
		}
	}
	return nodes, nil
}

// parseNode parses a node record and verifies its signature.
func parseNode(source string) (*enode.Node, error) {
	if strings.HasPrefix(source, "enode://") {
		return enode.ParseV4(source)
	}
	r, err := parseRecord(source)
	if err != nil {
		return nil, err
	}
	return enode.New(enode.ValidSchemes, r)
}

// pulled from enrcmd.go in dsp2p cli library.
// parseRecord parses a node record from hex, base64, or raw binary input.
func parseRecord(source string) (*enr.Record, error) {
	bin := []byte(source)
	if d, ok := decodeRecordHex(bytes.TrimSpace(bin)); ok {
		bin = d
	} else if d, ok := decodeRecordBase64(bytes.TrimSpace(bin)); ok {
		bin = d
	}
	var r enr.Record
	err := rlp.DecodeBytes(bin, &r)
	return &r, err
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

// decodeRecordBase64 decodes a base64-encoded node record.
func decodeRecordBase64(b []byte) ([]byte, bool) {
	if bytes.HasPrefix(b, []byte("enr:")) {
		b = b[4:]
	}
	dec := make([]byte, base64.RawURLEncoding.DecodedLen(len(b)))
	n, err := base64.RawURLEncoding.Decode(dec, b)
	return dec[:n], err == nil
}

// getNeighbors returns the neighbors of a node.
func getNeighbors(enr string) []enode.Node {
	neighbors := []enode.Node{}
	fmt.Printf("Getting neighbors for %s\n", enr)
	// Take in ENR string, and parse it into a node object.
	TargetNode, err := enode.Parse(enode.ValidSchemes, enr)
	if err != nil {
		log.Fatalf("Failed to parse ENR record: %s", err.Error())
	}
	disc, _, err := startV4("", TargetNode.String(), "", "")
	if err != nil {
		log.Fatalf("Failed to start ephemeral discovery node: %s", err.Error())
	}
	defer disc.Close()
	// Find the neighbors of the target node.
	enodes := disc.LookupPubkey(TargetNode.Pubkey())
	for _, enode := range enodes {
		fmt.Println(enode.String())
		neighbors = append(neighbors, *enode)
	}
	return neighbors
}

func clearScreen() string {
	return "\033[H\033[2J"
}
