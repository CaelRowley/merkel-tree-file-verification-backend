package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"sort"
)

type Node struct {
	Hash  []byte
	Left  *Node
	Right *Node
}

func main() {
	allHashes := getTestFileHashes()

	root := buildMerkelTree(allHashes)
	fmt.Println(root)
}

func buildMerkelTree(hashes [][]byte) *Node {
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

func removeDir(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Println(err)
	}
}

func makeDir(path string) {
	err := os.Mkdir(path, 0755)
	if err != nil {
		fmt.Println(err)
	}
}

func writeDummyFiles(path string, amount int) {
	for i := 0; i < amount; i++ {
		fileName := fmt.Sprintf("%d.txt", i)
		fileContent := fmt.Sprintf("Hello %d", i)
		writeFile(path, fileName, fileContent)
	}
}

func writeFile(path string, name string, content string) {
	file, err := os.Create(path + "/" + name)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func getFileHash(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256([]byte(data))
	return hash[:], nil
}

func getTestFileHashes() [][]byte {
	path := "files"
	removeDir(path)
	makeDir(path)
	writeDummyFiles(path, 1000)

	pageSize := 1024
	dir, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	var allFiles []os.DirEntry
	for {
		files, err := dir.ReadDir(pageSize)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Fatal(err)
		}

		if len(files) == 0 {
			break
		}

		allFiles = append(allFiles, files...)
	}

	sort.Slice(allFiles, func(i, j int) bool {
		return allFiles[i].Name() < allFiles[j].Name()
	})

	var allHashes [][]byte
	for _, file := range allFiles {
		fileHash, err := getFileHash(path + "/" + file.Name())
		if err != nil {
			log.Fatal(err)
		}
		allHashes = append(allHashes, fileHash)
	}

	return allHashes
}
