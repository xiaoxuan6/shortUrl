package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/tidwall/gjson"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	r := gin.Default()
	r.GET("/favicon.ico", func(c *gin.Context) { c.String(http.StatusOK, "") })
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.POST("/create", ReferrerMiddleware(), func(c *gin.Context) {
		uri := c.PostForm("url")
		shortUrl := generateShortUrl(base64.StdEncoding.EncodeToString([]byte(uri)))
		_ = saveData(uri, shortUrl)

		c.JSON(http.StatusOK, gin.H{
			"message": "create successful!",
			"data":    shortUrl,
		})
	})
	r.NoRoute(func(c *gin.Context) {
		redirectUrl := searchShort(strings.TrimLeft(c.Request.URL.Path, "/"))
		if redirectUrl == "" {
			c.JSON(http.StatusOK, gin.H{
				"message": fmt.Sprintf("no such url: %s", c.Request.URL.Path),
			})
		}

		c.Redirect(http.StatusMovedPermanently, redirectUrl)
	})
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Path %s Method %s not exist", c.Request.URL.Path, c.Request.Method),
		})
	})
	_ = r.Run(":80")
}

func ReferrerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		referer := c.Request.Header.Get("Referer")
		if referer == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "short url not exist！",
			})
			c.Abort()
			return
		}

		u, err := url.Parse(referer)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "invalid referer",
			})
			c.Abort()
			return
		}

		_ = godotenv.Load()
		if strings.ToLower(u.Hostname()) != os.Getenv("ALLOW_HOST") {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "referer not allowed",
				"referer": referer,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func generateShortUrl(uri string) string {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// 这里是表单字段，不是文件上传
	_ = writer.WriteField("longUrl", uri)
	_ = writer.WriteField("shortKey", "")

	if err := writer.Close(); err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(http.MethodPost, "https://v1.mk/short", &body)
	req.Header.Set("content-type", "multipart/form-data; boundary="+writer.Boundary())
	req.Header.Set("origin", "https://v1.mk")
	req.Header.Set("referer", "https://v1.mk/")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/155.0.0.0 Safari/537.36")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	response, _ := ioutil.ReadAll(resp.Body)
	shortUrl := gjson.GetBytes(response, "ShortUrl").String()
	return strings.ReplaceAll(shortUrl, "https://v1.mk/", "")
}

var (
	db *gorm.DB
)

func Init() {
	dns := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"),
	)

	_db, err := gorm.Open(mysql.Open(dns), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("数据库链接失败：%s", err.Error()))
	}

	db = _db
}

func AutoMigrate() {
	db.Logger = logger.Default.LogMode(logger.Silent) // 关闭慢查询日志
	_ = db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(Short{})
	_ = db.Migrator().CreateIndex(&Short{}, "short_index")
}

func Close() error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	err = sqlDB.Close()
	if err != nil {
		return err
	}

	return nil
}

func load() {
	_ = godotenv.Load()
	Init()
	//defer func() { _ = Close() }()
	AutoMigrate()
}

type Short struct {
	gorm.Model
	Uri    string `json:"uri" gorm:"varchar(255);index:uri,unique;not null;comment:'uri 地址'"`
	Target string `json:"target" gorm:"varchar(255);not null;comment:'短链目标'"`
}

func saveData(uri, target string) error {
	load()

	short := &Short{
		Uri:    uri,
		Target: target,
	}

	return db.Model(&Short{}).Create(short).Error
}

func searchShort(target string) string {
	load()

	println(target)
	short := Short{}
	if err := db.Model(&Short{}).Where("target = ?", target).First(&short).Error; err != nil {
		println(err.Error())
		return ""
	}

	return short.Uri
}
