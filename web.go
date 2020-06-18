package main

import (
	"bufio"
	"bytes"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func runServer(listenAt string) {
	r := gin.Default()

	r.Use(MaxAllowed(3))

	// r.POST("/s", query)
	r.GET("/admin/reload", reload)
	r.GET("/pic", getPicture)
	r.Run(listenAt)
}

func reload(ctx *gin.Context) {
	go reloadData()
	ctx.JSON(200, gin.H{
		"code": 0,
		"msg":  "ok",
	})
}

func getPicture(ctx *gin.Context) {
	log.Println("1", time.Now())
	// digest := ctx.Query("id")
	// if digest == "" {
	// 	ctx.JSON(401, gin.H{
	// 		"code": 401,
	// 		"msg":  "missing args",
	// 	})
	// 	return
	// }
	asfile := false
	if ctx.Query("asfile") == "1" {
		asfile = true
	}
	size, sSize := 0, ctx.Query("size")
	if sSize != "" {
		i64Size, err := strconv.ParseInt(sSize, 10, 32)
		if err != nil {
			ctx.JSON(401, gin.H{
				"code": 401,
				"msg":  err.Error(),
			})
			return
		}
		size = int(i64Size)
	}

	type Result struct {
		Path string
		Name string
		Ext  string
	}
	rows := []Result{}
	db.Raw(
		`SELECT full_path as path, resources.digest as name, resources.extname as ext
		FROM resource_property
		INNER JOIN resources ON resource_property.resource_id=resources.id
		WHERE used=(select min(used) from resource_property)
		ORDER BY resources.digest
		LIMIT 100`).Scan(&rows)

	n := len(rows)
	if n == 0 {
		ctx.JSON(404, gin.H{
			"code": 404,
			"msg":  "not found",
		})
		return
	}

	log.Println("2", time.Now())
	seq := rand.Intn(len(rows)-1) + 1
	pic := rows[seq]
	defer db.Model(&ResourceProperty{}).Where(&ResourceProperty{FullPath: pic.Path}).UpdateColumn("used", gorm.Expr("used+1"))
	log.Println("3", time.Now())
	// pic := queryResourceByDigest(conn, digest)
	// if pic == nil {
	// 	ctx.JSON(404, gin.H{
	// 		"code": 404,
	// 		"msg":  "not found",
	// 	})
	// 	return
	// }

	filename := pic.Name + "." + pic.Ext
	format, _ := imaging.FormatFromFilename(filename)
	img, _ := imaging.Open(pic.Path)
	x, y := img.Bounds().Max.X, img.Bounds().Max.Y

	max := func(x, y int) int {
		if x > y {
			return x
		}
		return y
	}

	if size != 0 && max(x, y) > size {
		ratio := float64(max(x, y)) / float64(size)
		xPixel, yPixel := int(float64(x)/ratio), int(float64(y)/ratio)
		img = imaging.Resize(img, xPixel, yPixel, imaging.Lanczos)
	}

	log.Println("4", time.Now())
	buffer := bytes.Buffer{}
	wr := bufio.NewWriter(&buffer)
	imaging.Encode(wr, img, format, imaging.JPEGQuality(90))

	log.Println("5", time.Now())

	w := ctx.Writer
	header := w.Header()

	header.Set("Content-Type", mime.TypeByExtension(filepath.Ext(filename)))
	w.WriteHeader(http.StatusOK)
	if asfile {
		// add header for download file
		w.Header().Set("content-disposition", "attachment; filename=\""+filename+"\"")
	}
	w.Write(buffer.Bytes())
	w.(http.Flusher).Flush()
}
