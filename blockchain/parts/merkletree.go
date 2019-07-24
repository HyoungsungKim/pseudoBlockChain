package parts

import (
	"crypto/sha256"
)

//MerkleTree type means merkle tree
type MerkleTree struct {
	RootNode *MerkleNode
}

//MerkleNode is one of merkletree's node
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

//NewMerkleNode make new merkle node
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	mNode := MerkleNode{}

	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		mNode.Data = hash[:]
	} else {
		prevHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashes)
		mNode.Data = hash[:]
	}

	mNode.Left = left
	mNode.Right = right

	return &mNode
}

//NewMerkleTree generate merkle tree
//Good way to implement tree from bottom to top
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	//padding
	//Padded data is same date with last node
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, datum := range data {
		//Serialize merkle node
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}

	//I think log(data) is enough
	for i := 0; i < len(data)/2; i++ {
		var newLevel []MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			newLevel = append(newLevel, *node)
		}

		nodes = newLevel
	}

	mTree := MerkleTree{&nodes[0]}

	return &mTree
}
