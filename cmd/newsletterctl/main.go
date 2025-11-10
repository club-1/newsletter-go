package main

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
)

const Name = "newsletterctl"

func Subscribe() {
	log.Println("recieved mail to route 'subscribe'")
}

func SubscribeConfirm() {
	log.Println("recieved mail to route 'subscribe-confirm'")
}

func Unsubscribe() {
	log.Println("recieved mail to route 'unsubscribe'")
}

func Send() {
	log.Println("recieved mail to route 'send'")
}

func SendConfirm() {
	log.Println("recieved mail to route 'send-confirm'")
}

func initLogger() *os.File {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalln("cannot get user cache directory:", err)
	}

	logDir := filepath.Join(userCacheDir, "newsletter")

	err = os.MkdirAll(logDir, 0775)
	if err != nil {
		log.Fatalln("cannot create log folder:", err)
	}
	LogFilePath := filepath.Join(logDir, Name+".log")

	logFile, err := os.OpenFile(LogFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
	if err != nil {
		panic(err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	return logFile
}

func main() {
	logFile := initLogger()
	defer logFile.Close()
	flag.Parse()

	args := flag.Args()

	if len(args) >= 1 {
		switch args[0] {
		case "subscribe":
			Subscribe()
		case "subscribe-confirm":
			SubscribeConfirm()
		case "unsubscribe":
			Unsubscribe()
		case "send":
			Send()
		case "send-confirm":
			SendConfirm()
		default:
			log.Fatalln("invalid sub command:", args[0])
		}
	}
}
