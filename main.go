package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	db *gorm.DB
)

func reloadData() {
	log.Println("reload ...")
	defer log.Println("reload finish")

	paths := []Path{}
	db.Find(&paths)
	for _, path := range paths {
		files := walk(path.Path, path.Recursive, func(name string) bool {
			name = strings.ToLower(name)
			if strings.HasSuffix(name, ".png") {
				return true
			}
			if strings.HasSuffix(name, ".jpg") {
				return true
			}
			if strings.HasSuffix(name, ".jpeg") {
				return true
			}
			return false
		})

		for _, filepath := range files {
			if resouceExists(filepath) {
				continue
			}
			addResource(filepath, path.ID)
		}
	}

	files := []ResourceProperty{}
	db.Find(&files)
	for _, item := range files {
		_, err := os.Stat(item.FullPath)
		if os.IsNotExist(err) {
			db.Debug().Unscoped().Delete(&ResourceProperty{FullPath: item.FullPath})
		}
	}
}

func resouceExists(fullpath string) bool {
	count := 0
	db.Model(&ResourceProperty{}).Where(&ResourceProperty{FullPath: fullpath}).Count(&count)
	if count > 0 {
		return true
	}
	return false
}

func addResource(fullpath string, pathID uint64) {
	digest := sha1sum(fullpath)
	if digest == "" {
		return
	}
	extname := filepath.Ext(fullpath)
	if strings.HasPrefix(extname, ".") {
		extname = extname[1:]
	}

	resource := Resources{
		Digest:  digest,
		Extname: extname,
		Size:    0,
	}
	result := db.Where(&resource).FirstOrCreate(&resource)
	if result.Error != nil {
		log.Println(result.Error)
		return
	}

	property := ResourceProperty{
		ResourceID: resource.ID,
		FullPath:   fullpath,
		PathID:     pathID,
	}
	result = db.Where(&property).FirstOrCreate(&property)
	if result.Error != nil {
		log.Println(result.Error)
		return
	}
	log.Println("add file:", fullpath, resource.ID)
}

func main() {
	config := initConfig()

	db = initDB(config.Dsn)
	if db == nil {
		return
	}
	defer db.Close()

	go reloadData()

	listenAddr := "127.0.0.1:8081"
	if config.Addr != "" {
		listenAddr = config.Addr
	}
	runServer(listenAddr)
}
