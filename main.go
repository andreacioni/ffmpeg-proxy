package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/andreacioni/ffmpeg-proxy/config"
	"github.com/andreacioni/ffmpeg-proxy/ffmpeg"
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("invalid number of arguments, missing config file path")
	}

	cfg, err := config.Load(os.Args[1])
	if err != nil {
		log.Fatalf("failed to load config file: %+v\n", err)
	}

	ffmpeg.Init(cfg.Ffmpeg)

	resetTickChh := make(chan bool)
	stopChn := make(chan bool)

	go func() {
		lastTick := time.Now().Unix()
		for {
			select {
			case <-resetTickChh:
				lastTick = time.Now().Unix()
			case <-stopChn:
				ffmpeg.Stop()
				return
			default:
				if time.Now().Unix() > lastTick+cfg.AutoStopAfter {
					if ffmpeg.IsRunning() {
						ffmpeg.Stop()
					}
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()

	router := gin.Default()

	srv := endless.NewServer(fmt.Sprintf(":%d", cfg.Port), router)

	router.GET("/*path", func(c *gin.Context) {
		go func() {
			resetTickChh <- true
		}()

		filename := filepath.Join(cfg.ServePath, c.Param("path"))
		ext := filepath.Ext(filename)

		if ext == ".m3u8" {
			if err := ffmpeg.Start(); err != nil {
				c.JSON(500, fmt.Sprintf("failed to start ffmpeg: %+v", err))
				return
			}
		}

		c.Header("X-File-Extension", ext)
		c.Header("Content-Type", contentTypeMap(ext))
		c.File(filename)
	})

	srv.SignalHooks[endless.PRE_SIGNAL][syscall.SIGINT] = append(
		srv.SignalHooks[endless.PRE_SIGNAL][syscall.SIGINT],
		shutdownHook)

	if err := srv.ListenAndServe(); err != nil {
		log.Printf("failed to start server: %+v\n", err)
	}
}

func shutdownHook() {
	log.Print("shutdown hook...")
	if err := ffmpeg.Close(); err != nil {
		log.Println("error closing ffmpeg: %w\n", err)
	} else {
		log.Println("done!")
	}
}

func contentTypeMap(ext string) string {
	if ext == ".m3u8" {
		return "audio/x-mpegur"
	} else if ext == ".ts" {
		return "video/mp2t"
	} else if ext == ".html" {
		return "text/html"
	}

	return "text/plain"
}
