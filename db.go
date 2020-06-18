package main

import (
	"log"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

// Resources is uniq picture with sha1
type Resources struct {
	ID        uint64     `gorm:"primary_key"`
	Digest    string     `gorm:"type:varchar(128);unique_index"`
	Extname   string     `gorm:"type:varchar(16);default:''"`
	Size      uint       `gorm:"default:0"`
	CreatedAt *time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func (Resources) TableName() string { return "resources" }

type ResourceProperty struct {
	ResourceID uint64
	PathID     uint64
	FullPath   string     `gorm:"type:varchar(512);unique_index"`
	Used       uint       `gorm:"default:0"`
	UpdateAt   *time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	CreatedAt  *time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func (ResourceProperty) TableName() string { return "resource_property" }

type Path struct {
	ID        uint64     `gorm:"primary_key"`
	Path      string     `gorm:"type:varchar(512);unique_index"`
	Ignore    string     `gorm:"type:varchar(128);default:''"`
	Recursive bool       `gorm:"default:false"`
	Label     string     `gorm:"type:varchar(128);default:''"`
	CreatedAt *time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func (Path) TableName() string { return "path" }

func initDB(dsn string) *gorm.DB {
	dsn += "?parseTime=true"
	s := strings.Split(dsn, "://")
	db, err := gorm.Open(s[0], s[1])
	if err != nil {
		log.Println(err)
		return nil
	}

	if !db.HasTable(&Path{}) {
		db.CreateTable(&Path{})
	}
	if !db.HasTable(&Resources{}) {
		db.CreateTable(&Resources{})
	}
	if !db.HasTable(&ResourceProperty{}) {
		db.CreateTable(&ResourceProperty{})
	}
	db.AutoMigrate(&Path{}, &Resources{}, &ResourceProperty{})

	return db
}
