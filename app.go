package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/parnurzeal/gorequest"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

var (
	conn *gorm.DB
	cx   string
	key  string
)

type Config struct {
	Key  string `yaml:"key"`
	Cx   string `yaml:"cx"`
	Dsn  string `yaml:"db_dsn"`
	Addr string `yaml:"listen"`
}

type SearchHistory struct {
	ID        uint       `gorm:"primary_key"`
	Keyword   string     `gorm:"type:varchar(128)"`
	Start     uint       `gorm:"default:0"`
	End       uint       `gorm:"default:0"`
	Total     uint       `gorm:"default:0"`
	Result    string     `gorm:"type:text"`
	CreatedAt *time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func (SearchHistory) TableName() string { return "search_history" }

func addSearchHistory(db *gorm.DB, keyword string, start, end, total int, result string) int {
	item := SearchHistory{
		Keyword: keyword,
		Start:   uint(start),
		End:     uint(end),
		Total:   uint(total),
		Result:  result,
	}
	db.Create(&item)
	return int(item.ID)
}

func querySearchHistory(db *gorm.DB, keyword string) *SearchHistory {
	item := SearchHistory{}
	db.Where(&SearchHistory{Keyword: keyword}).Find(&item)
	return &item
}

type Resources struct {
	ID        uint       `gorm:"primary_key"`
	URL       string     `gorm:"type:varchar(512)"`
	Digest    string     `gorm:"type:varchar(128);unique_index"`
	Extname   string     `gorm:"type:varchar(16);default:''"`
	SearchID  uint       `gorm:"default:0"`
	Used      uint       `gorm:"default:0"`
	CreatedAt *time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func addResource(db *gorm.DB, url, digest, extname string, searchID int) int {
	item := Resources{
		URL:      url,
		Digest:   digest,
		Extname:  extname,
		SearchID: uint(searchID),
	}
	db.Create(&item)
	return int(item.ID)
}

func queryResourceByKeyword(db *gorm.DB, keyword string) []Resources {
	search := SearchHistory{}
	result := db.Where(&SearchHistory{Keyword: keyword}).First(&search)
	if result.Error == gorm.ErrRecordNotFound {
		println("not found")
		return nil
	}
	resources := []Resources{}
	db.Where(&Resources{SearchID: search.ID}).Find(&resources)
	return resources
}

func queryResourceByDigest(db *gorm.DB, digest string) *Resources {
	resource := Resources{}
	result := db.Where(&Resources{Digest: digest}).First(&resource)
	if result.Error == gorm.ErrRecordNotFound {
		println("not found")
		return nil
	}
	return &resource
}

func downloadPics(url, path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}

	req := gorequest.New().Timeout(8 * time.Second)
	req.Get(url)
	req.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.97 Safari/537.36")
	resp, _, errs := req.End()
	if len(errs) > 0 {
		println(errs)
		return ""
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err.Error())
		return ""
	}
	mimeType := resp.Header.Get("Content-Type")
	extnames, _ := mime.ExtensionsByType(mimeType)
	extname := extnames[0]
	hexDigest := md5.Sum(data)
	digest := hex.EncodeToString(hexDigest[:])

	filename := digest + extname
	f, err := os.Create(path + "/" + filename)
	if err != nil {
		println(err.Error())
		return ""
	}
	f.Write(data)
	f.Close()
	return filename
}

func query(ctx *gin.Context) {
	secret, keyword := ctx.PostForm("k"), ctx.PostForm("q")
	if secret != "123321" || keyword == "" {
		ctx.JSON(401, gin.H{
			"code": 401,
			"msg":  "missing args",
		})
		return
	}

	pics := queryResourceByKeyword(conn, keyword)
	if len(pics) > 0 {
		md5s := []string{}
		for _, pic := range pics {
			md5s = append(md5s, pic.Digest)
		}
		ctx.JSON(200, gin.H{
			"code": 0,
			"msg":  md5s,
		})
		return
	}

	req := gorequest.New().Timeout(8 * time.Second)
	req.Get("https://www.googleapis.com/customsearch/v1")
	req.Param("key", key)
	req.Param("cx", cx)
	req.Param("searchType", "image")
	req.Param("num", "10")
	req.Param("q", keyword)

	resp, _, errs := req.End()
	if len(errs) > 0 {
		println(errs)
		ctx.JSON(500, gin.H{
			"code": 401,
			"msg":  errs,
		})
		return
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	body := string(data)
	start := gjson.Get(body, "queries.request.0.startIndex")
	count := gjson.Get(body, "queries.request.0.count")
	total := gjson.Get(body, "queries.request.0.totalResults")
	searchID := addSearchHistory(
		conn,
		keyword,
		int(start.Int()),
		int(count.Int()),
		int(total.Int()),
		body)

	md5s := []string{}
	items := gjson.Get(body, "items").Array()
	for _, item := range items {
		link := item.Get("link").String()
		filename := downloadPics(link, "download")
		if filename == "" {
			continue
		}
		s := strings.Split(filename, ".")
		addResource(conn, link, s[0], s[1], searchID)
		md5s = append(md5s, s[0])
	}

	ctx.JSON(200, gin.H{
		"code": 0,
		"msg":  md5s,
	})
}

func getPicture(ctx *gin.Context) {
	digest := ctx.Query("id")
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

	pic := queryResourceByDigest(conn, digest)
	if pic == nil {
		ctx.JSON(404, gin.H{
			"code": 404,
			"msg":  "not found",
		})
		return
	}

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

func initConfig() Config {
	cfgFile, err := os.Open("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer cfgFile.Close()

	config := Config{}
	cfg := yaml.NewDecoder(cfgFile)
	err = cfg.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

func main() {
	config := initConfig()
	cx = config.Cx
	key = config.Key
	listenAddr := "127.0.0.1:8081"
	if config.Addr != "" {
		listenAddr = config.Addr
	}

	dsn := config.Dsn + "?parseTime=true"
	s := strings.Split(dsn, "://")
	println(dsn)
	db, err := gorm.Open(s[0], s[1])
	if err != nil {
		println(err.Error())
		return
	}
	conn = db
	defer db.Close()
	db.AutoMigrate(&SearchHistory{}, &Resources{})

	r := gin.Default()
	r.POST("/s", query)
	r.GET("/pic", getPicture)
	r.Run(listenAddr)
}
