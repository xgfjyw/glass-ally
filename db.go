package main

import (
	"time"
)

// type SearchHistory struct {
// 	ID        uint       `gorm:"primary_key"`
// 	Keyword   string     `gorm:"type:varchar(128)"`
// 	Start     uint       `gorm:"default:0"`
// 	End       uint       `gorm:"default:0"`
// 	Total     uint       `gorm:"default:0"`
// 	Result    string     `gorm:"type:text"`
// 	CreatedAt *time.Time `gorm:"default:CURRENT_TIMESTAMP"`
// }

// func (SearchHistory) TableName() string { return "search_history" }

// func addSearchHistory(db *gorm.DB, keyword string, start, end, total int, result string) int {
// 	item := SearchHistory{
// 		Keyword: keyword,
// 		Start:   uint(start),
// 		End:     uint(end),
// 		Total:   uint(total),
// 		Result:  result,
// 	}
// 	db.Create(&item)
// 	return int(item.ID)
// }

// func querySearchHistory(db *gorm.DB, keyword string) *SearchHistory {
// 	item := SearchHistory{}
// 	db.Where(&SearchHistory{Keyword: keyword}).Find(&item)
// 	return &item
// }

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

// func addResource(db *gorm.DB, url, digest, extname string, searchID int) int {
// 	item := Resources{
// 		URL:      url,
// 		Digest:   digest,
// 		Extname:  extname,
// 		SearchID: uint(searchID),
// 	}
// 	db.Create(&item)
// 	return int(item.ID)
// }

// func queryResourceByKeyword(db *gorm.DB, keyword string) []Resources {
// 	search := SearchHistory{}
// 	result := db.Where(&SearchHistory{Keyword: keyword}).First(&search)
// 	if result.Error == gorm.ErrRecordNotFound {
// 		println("not found")
// 		return nil
// 	}
// 	resources := []Resources{}
// 	db.Where(&Resources{SearchID: search.ID}).Find(&resources)
// 	return resources
// }

// func queryResourceByDigest(db *gorm.DB, digest string) *Resources {
// 	resource := Resources{}
// 	result := db.Where(&Resources{Digest: digest}).First(&resource)
// 	if result.Error == gorm.ErrRecordNotFound {
// 		println("not found")
// 		return nil
// 	}
// 	return &resource
// }

// func incrResourceUse(db *gorm.DB, digest string) *Resources {
// 	resource := Resources{Digest: digest}
// 	db.Find(&resource).UpdateColumn("used", gorm.Expr("used+1"))
// 	return &resource
// }

// func (SearchHistory) TableName() string { return "search_history" }
