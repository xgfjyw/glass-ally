package main

import (
	"bufio"
	"bytes"
	"math/rand"
	"mime"
	"net/http"
	"strconv"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

func runServer(listenAt string) {
	r := gin.Default()
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

	resources := []Resources{}
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

	filename := pic.Digest + "." + pic.Extname
	format, _ := imaging.FormatFromFilename(filename)
	img, _ := imaging.Open("download/" + filename)
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

	header.Set("Content-Type", mime.TypeByExtension("."+pic.Extname))
	w.WriteHeader(http.StatusOK)
	w.Write(buffer.Bytes())
	w.(http.Flusher).Flush()
}
