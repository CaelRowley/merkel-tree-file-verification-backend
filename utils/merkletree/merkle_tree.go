package merkletree

import (
	"crypto/sha256"

	"github.com/google/uuid"
)

var Trees []MerkleTree

type MerkleTree struct {
	ID   uuid.UUID
	Root *Node
}

type Node struct {
	Hash  []byte
	Left  *Node
	Right *Node
}

func BuildTree(hashes [][]byte) *Node {
	var currentLevel []*Node

	for _, hash := range hashes {
		currentLevel = append(currentLevel, newNode(hash, nil, nil))
	}

	// If len(currentLevel) is 1 we are at the root
	for len(currentLevel) > 1 {
		var nextLevel []*Node

		// Is a binary tree so each node must have two children
		if len(currentLevel)%2 != 0 {
			currentLevel = append(currentLevel, currentLevel[len(currentLevel)-1])
		}

		for i := 0; i < len(currentLevel); i += 2 {
			left := currentLevel[i]
			right := currentLevel[i+1]
			hash := hashPair(left.Hash, right.Hash)

			newNode := newNode(hash, left, right)
			nextLevel = append(nextLevel, newNode)
		}

		currentLevel = nextLevel
	}

	return currentLevel[0]
}

func AddTree(tree MerkleTree) {
	Trees = append(Trees, tree)
}

func newNode(hash []byte, left *Node, right *Node) *Node {
	return &Node{
		Hash:  hash,
		Left:  left,
		Right: right,
	}
}

func hashPair(left []byte, right []byte) []byte {
	pair := append(left, right...)
	hash := sha256.Sum256(pair)
	return hash[:]
}
