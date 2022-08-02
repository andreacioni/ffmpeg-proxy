package ffmpeg

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/andreacioni/ffmpeg-proxy/utils"
)

const waitTimeout = 2

var ffmpegChnCmd = make(chan bool)
var ffmpegChnShutdown = make(chan bool)
var ffmpegChnState = make(chan error)

var command *exec.Cmd
var outputDirectory string

func Init(basePath string) {
	outputDirectory = basePath
	go ffmpeg()
}

func Start() error {
	ffmpegChnCmd <- true

	if err := <-ffmpegChnState; err != nil {
		return fmt.Errorf("ffmpeg failed to start: %w", err)
	}

	return nil
}

func Stop() {
	ffmpegChnCmd <- false
}

func Close() {
	ffmpegChnShutdown <- true
}

func ffmpegRunning() bool {
	return command != nil && command.Process != nil
}

func ffmpegKill() error {
	if !ffmpegRunning() {
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
	filename := filepath.Join(outputDirectory, "file.m3u8")

	if ffmpegRunning() {
		log.Println("[ffmpeg] already running")
		return nil
	}

	if utils.FileExists(filename) {
		log.Printf("[ffmpeg] %s is present\n", filename)
		if err := os.Remove(filename); err != nil {
			log.Printf("[ffmpeg] %s can't be deleted\n", filename)
			return err
		}
		log.Printf("[ffmpeg] %s deleted\n", filename)
	}

	command = exec.Command("touch", filename)

	if err := command.Start(); err != nil {
		return err
	}

	waitEnd := time.Now().Unix() + waitTimeout
	for time.Now().Unix() < waitEnd {
		if utils.FileExists(filename) {
			log.Printf("[ffmpeg] %s created, process is running\n", filename)
			return nil
		}
	}

	return fmt.Errorf("%s not found", filename)
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
