package live

import (
	"fmt"
	"nsy_chat_live/config"
	"os"
	"os/exec"
	"strings"
	"time"
)

var ffmpegPath string

func initFfmpeg() {
	if _, err := os.Stat(config.Conf.FfmpegPath); err == nil {
		ffmpegPath = config.Conf.FfmpegPath
	} else if _, err := os.Stat("./ffmpeg.exe"); err == nil {
		ffmpegPath = "./ffmpeg.exe"
	} else {
		ffmpegPath = "ffmpeg.exe"
	}
}

func startFfmpegWatcher() {
	initFfmpeg()
	fmt.Println("ffmpeg path: " + ffmpegPath + ", Starting ffmpeg watcher...")
	go func() {
		for nsyLiveInfo := range liveRecordChannel {
			if err := startFfmpegRecord(nsyLiveInfo); err != nil {
				fmt.Println("Error starting ffmpeg record: ", err)
				liveRecordChannel <- nsyLiveInfo
			}
		}
	}()
}

func startFfmpegRecord(nsyLiveInfo *NsyLiveInfo) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in ffmpeg record: ", r)
		}
	}()
	outputFile := fmt.Sprintf("%s_%s.mp4",
		nsyLiveInfo.Name, time.Now().Format("20060102"))

	logFileName := strings.TrimSuffix(outputFile, ".mp4") + ".txt"
	logFile, err := os.Create(logFileName)
	if err != nil {
		return fmt.Errorf("创建日志文件失败: %v", err)
	}

	cmd := exec.Command(
		//"cmd", "/C", "start", "cmd", "/K",
		ffmpegPath,
		"-i", nsyLiveInfo.RtmpUrl,
		"-c", "copy",
		outputFile,
	)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in ffmpeg record: ", r)
			}
		}()
		defer logFile.Close()
		fmt.Println("start recording by command: " + cmd.String())
		fmt.Println("ffmpeg 日志: " + logFile.Name())
		cmd.Stdout = logFile
		cmd.Stderr = logFile

		if err := cmd.Start(); err != nil {
			fmt.Println("ffmpeg err: " + err.Error())
		}

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop() // 确保定时器资源释放

		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()

		startMsg := fmt.Sprintf("[%s] 录制开始\n", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Print(startMsg)

		// 循环监控进程状态
		for {
			select {
			case <-ticker.C:
				runningMsg := fmt.Sprintf("[%s] 录制中...日志: %s\n", time.Now().Format("2006-01-02 15:04:05"), logFile.Name())
				fmt.Print(runningMsg)
			case err := <-done: // 进程结束处理
				endMsg := fmt.Sprintf("[%s] Done: 录制结束\n", time.Now().Format("2006-01-02 15:04:05"))
				fmt.Print(endMsg)
				logFile.WriteString(endMsg)

				if err != nil {
					errMsg := fmt.Sprintf("[%s] 录制错误: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
					fmt.Print(errMsg)
				}
				return
			}
		}
	}()
	return nil
}
