package fileutil

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"sort"
)

type File struct {
	Name string
	Data []byte
}

func GetTestFileHashes(path string) [][]byte {
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

	sort.Slice(allFiles, func(i int, j int) bool {
		return allFiles[i].Name() < allFiles[j].Name()
	})

	var allHashes [][]byte
	for _, file := range allFiles {
		fileHash, err := GetFileHash(path + "/" + file.Name())
		if err != nil {
			log.Fatal(err)
		}
		allHashes = append(allHashes, fileHash)
	}

	return allHashes
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

func GetFileHash(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256([]byte(data))
	return hash[:], nil
}

func GetFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func GetFiles(path string) []File {
	writeDummyFiles(path, 1000)

	pageSize := 1024
	dir, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	var allDirs []os.DirEntry
	for {
		dirs, err := dir.ReadDir(pageSize)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Fatal(err)
		}

		if len(dirs) == 0 {
			break
		}

		allDirs = append(allDirs, dirs...)
	}

	sort.Slice(allDirs, func(i int, j int) bool {
		return allDirs[i].Name() < allDirs[j].Name()
	})

	var allFiles []File
	for _, dir := range allDirs {
		file, err := GetFile(path + "/" + dir.Name())
		if err != nil {
			log.Fatal(err)
		}
		newFile := File{
			Name: dir.Name(),
			Data: file,
		}
		allFiles = append(allFiles, newFile)
	}

	return allFiles
}
