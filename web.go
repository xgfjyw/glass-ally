package main

import (
	"bufio"
	"bytes"
	"math/rand"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"

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
	// digest := ctx.Query("id")
	// if digest == "" {
	// 	ctx.JSON(401, gin.H{
	// 		"code": 401,
	// 		"msg":  "missing args",
	// 	})
	// 	return
	// }
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

	resources := []ResourceProperty{}
	db.Find(&resources)
	n := len(resources)
	if n == 0 {
		ctx.JSON(404, gin.H{
			"code": 404,
			"msg":  "not found",
		})
		return
	}

	seq := rand.Intn(n-1) + 1
	pic := resources[seq]

	// pic := queryResourceByDigest(conn, digest)
	// if pic == nil {
	// 	ctx.JSON(404, gin.H{
	// 		"code": 404,
	// 		"msg":  "not found",
	// 	})
	// 	return
	// }

	filename := pic.FullPath
	format, _ := imaging.FormatFromFilename(filename)
	img, _ := imaging.Open(filename)
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

	buffer := bytes.Buffer{}
	wr := bufio.NewWriter(&buffer)
	imaging.Encode(wr, img, format)

	w := ctx.Writer
	header := w.Header()

	header.Set("Content-Type", mime.TypeByExtension(filepath.Ext(filename)))
	w.WriteHeader(http.StatusOK)

	resource := Resources{}
	_, downloadFilename := filepath.Split(filename)
	result := db.Where(&Resources{ID: pic.ResourceID}).First(&resource)
	if result.Error != gorm.ErrRecordNotFound {
		downloadFilename = resource.Digest + "." + resource.Extname
	}
	w.Header().Set("content-disposition", "attachment; filename=\""+downloadFilename+"\"")
	w.Write(buffer.Bytes())
	w.(http.Flusher).Flush()
}
