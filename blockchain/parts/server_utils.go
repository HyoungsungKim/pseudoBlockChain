package parts

import "fmt"

const (
	protocol      = "tcp"
	nodeVersion   = 1
	commandLength = 12
)

var nodeAddress string
var miningAddress string
var knownNodes = []string{"localhost:3000"}
var blocksInTransit = [][]byte{}
var mempool = make(map[string]Transaction)

//verzion version is already declared
//verzion show information of node
type verzion struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

type addr struct {
	AddrList []string
}

type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type getblocks struct {
	AddrFrom string
}

type getdata struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type block struct {
	AddrFrom string
	Block    []byte
}

type tx struct {
	AddFrom     string
	Transaction []byte
}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

func extractCommand(request []byte) []byte {
	return request[:commandLength]
}
