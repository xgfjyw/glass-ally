package main

import (
	"bufio"
	"bytes"
	"log"
	"math"
	"math/rand"
	"net/http"
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
	start := time.Now()
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
	resolution, sResolution := 0, ctx.Query("size")
	if sResolution != "" {
		i64Resolution, err := strconv.ParseInt(sResolution, 10, 32)
		if err != nil {
			ctx.JSON(401, gin.H{
				"code": 401,
				"msg":  err.Error(),
			})
			return
		}
		resolution = int(i64Resolution)
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

	seq := rand.Intn(n) + 1
	pic := rows[seq]
	defer db.Model(&ResourceProperty{}).Where(&ResourceProperty{FullPath: pic.Path}).UpdateColumn("used", gorm.Expr("used+1"))
	log.Println("db", time.Now().Sub(start).Microseconds()/1000, "ms")

	img, err := imaging.Open(pic.Path)
	if err != nil {
		ctx.JSON(500, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	max := func(x, y int) int {
		if x > y {
			return x
		}
		return y
	}
	x, y := img.Bounds().Max.X, img.Bounds().Max.Y
	if resolution != 0 && max(x, y) > resolution {
		ratio := float64(max(x, y)) / float64(resolution)
		xPixel, yPixel := int(float64(x)/ratio), int(float64(y)/ratio)
		img = imaging.Resize(img, xPixel, yPixel, imaging.CatmullRom)
	}

	log.Println("open + resize", time.Now().Sub(start).Microseconds()/1000, "ms")

	expectedSize, minQuaity := 1024*168, 62
	buffer, quality := bytes.Buffer{}, 93
	for {
		wr := bufio.NewWriter(&buffer)
		imaging.Encode(wr, img, imaging.JPEG, imaging.JPEGQuality(quality))

		size := buffer.Len()
		if size <= expectedSize || quality <= minQuaity {
			break
		}

		delta := int(math.Pow(4.0*float64(size)/float64(expectedSize), 1.3))
		switch {
		case delta > 15:
			delta = 15
		case delta < 5:
			delta = 5
		}
		quality -= delta
		if quality < minQuaity {
			quality = minQuaity
		}
		buffer.Reset()
	}

	log.Println("encode", time.Now().Sub(start).Microseconds()/1000, "ms")

	w := ctx.Writer
	header := w.Header()

	header.Set("Content-Type", "image/jpeg")
	w.WriteHeader(http.StatusOK)
	if asfile {
		// add header for download file
		filename := pic.Name + "." + pic.Ext
		w.Header().Set("content-disposition", "attachment; filename=\""+filename+"\"")
	}
	w.Write(buffer.Bytes())
	w.(http.Flusher).Flush()
}
