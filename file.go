package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func walk(path string, recursive bool, filter func(string) bool) []string {
	if recursive {
		return walkRecursiveImp(path, filter)
	}
	return walkImp(path, filter)
}

func walkImp(path string, filter func(string) bool) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Println(err)
		return nil
	}

	result := []string{}
	for _, info := range files {
		if info.IsDir() == false && filter(info.Name()) {
			result = append(result, path+"/"+info.Name())
		}
	}
	return result
}

func walkRecursiveImp(path string, filter func(string) bool) []string {
	files := []string{}
	err := filepath.Walk(path,
		func(fullpath string, info os.FileInfo, err error) error {
			if err != nil {
				log.Println(err)
				return err
			}
			fmt.Println(fullpath, info.Size())

			if info.IsDir() == false && filter(info.Name()) {
				files = append(files, fullpath+"/"+info.Name())
			}

			return nil
		})
	if err != nil {
		log.Println(err)
	}

	return files
}

func sha1sum(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer file.Close()

	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Println(err)
		return ""
	}
	hexDigest := hash.Sum(nil)
	return hex.EncodeToString(hexDigest)
}
