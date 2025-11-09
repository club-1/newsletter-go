package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/club-1/newsletter-go"
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

func main() {
	prefix, err := getPrefix()
	if err != nil {
		l.Fatalln("cannot get prefix:", err)
	}
	mail := &newsletter.Mail{
		Body: "coucou",
	}
	l.Println("hello world:", mail, prefix)
}
