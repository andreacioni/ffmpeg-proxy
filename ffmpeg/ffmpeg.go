package ffmpeg

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/andreacioni/ffmpeg-proxy/config"
	"github.com/andreacioni/ffmpeg-proxy/utils"
)

var (
	ffmpegChnCmd      = make(chan bool)
	ffmpegChnShutdown = make(chan bool)
	ffmpegChnState    = make(chan error)
	command           *exec.Cmd
	cfg               config.FFmpegConfig
)

func Init(c config.FFmpegConfig) {
	cfg = c
	go ffmpeg()
}

func Start() error {
	ffmpegChnCmd <- true

	if err := <-ffmpegChnState; err != nil {
		return fmt.Errorf("ffmpeg failed to start: %w", err)
	}

	return nil
}

func Stop() error {
	ffmpegChnCmd <- false

	if err := <-ffmpegChnState; err != nil {
		return fmt.Errorf("ffmpeg failed to stop: %w", err)
	}

	return nil
}

func Close() {
	ffmpegChnShutdown <- true
}

func IsRunning() bool {
	return command != nil && command.Process != nil
}

func ffmpegKill() error {
	if !IsRunning() {
		log.Println("[ffmpeg] not running")
		return nil
	}

	if err := command.Process.Signal(os.Interrupt); err != nil {
		command = nil
		return err
	}

	if err := command.Wait(); err != nil {
		log.Printf("[ffmpeg] process return error: %+v", err)
	}
	log.Printf("[ffmpeg] process exits with code: %d\n", command.ProcessState.ExitCode())

	command = nil
	return nil
}

func ffmpegExec() error {
	if IsRunning() {
		log.Println("[ffmpeg] already running")
		return nil
	}

	if utils.FileExists(cfg.IndexFile) {
		log.Printf("[ffmpeg] %s is present\n", cfg.IndexFile)
		if err := os.Remove(cfg.IndexFile); err != nil {
			log.Printf("[ffmpeg] %s can't be deleted\n", cfg.IndexFile)
			return err
		}
		log.Printf("[ffmpeg] %s deleted\n", cfg.IndexFile)
	}

	command = exec.Command(cfg.Command, cfg.Args...)

	if err := command.Start(); err != nil {
		return err
	}

	waitEnd := time.Now().Unix() + cfg.WaitForIndex
	for time.Now().Unix() < waitEnd {
		if utils.FileExists(cfg.IndexFile) {
			log.Printf("[ffmpeg] %s created, process is running\n", cfg.IndexFile)
			return nil
		}
	}

	ffmpegKill()
	return fmt.Errorf("%s not found", cfg.IndexFile)
}

func ffmpeg() {
	log.Println("[ffmpeg] start")
	for {
		select {
		case <-ffmpegChnShutdown:
			log.Println("[ffmpeg] shuting down")
			ffmpegKill()
			return
		case s := <-ffmpegChnCmd:
			log.Printf("[ffmpeg] received command: %t\n", s)
			if s {
				ffmpegChnState <- ffmpegExec()
			} else {
				ffmpegChnState <- ffmpegKill()
			}
		}
	}
}
