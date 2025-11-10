package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const Name = "newsletter"

var l = log.New(os.Stderr, Name+": ", 0)

func getPrefix() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("get executable path: %w", err)
	}
	realpath, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return "", fmt.Errorf("eval symlinks: %w", err)
	}
	return filepath.Dir(filepath.Dir(realpath)), nil
}

func Init() {
	prefix, err := getPrefix()
	if err != nil {
		l.Fatalln("cannot get prefix:", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		l.Fatalln("cannot get user home directory:", err)
	}
	routes := [5]string{"subscribe", "unsubscribe", "subscribe-confirm", "send", "send-confirm"}
	for _, route := range routes {
		fileName := ".forward+" + route
		filePath := filepath.Join(homeDir, fileName)
		l.Println("writting file", filePath)

		cmdPath := filepath.Join(prefix, "sbin/newsletterctl")
		content := []byte("| \"" + cmdPath + " " + route + "\"\n")
		err := os.WriteFile(filePath, content, 0664)
		if err != nil {
			l.Println("cannot write file:", filePath)
		}
	}
}

func main() {
	log.SetFlags(0)

	flag.Parse()

	args := flag.Args()

	if len(args) >= 1 {
		switch args[0] {
		case "init":
			Init()
		default:
			log.Fatalln("invalid sub command")
		}
	}
}
