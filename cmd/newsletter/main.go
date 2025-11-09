package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/club-1/newsletter-go"
)

const Name = "newsletter"

var l = log.New(os.Stderr, Name, 0)

func main() {
	executable, err := os.Executable()
	if err != nil {
		l.Fatalln("cannot get executable path:", err)
	}
	prefix := filepath.Dir(filepath.Dir(executable))
	mail := &newsletter.Mail{
		Body: "coucou",
	}
	l.Println("hello world:", mail, prefix)
}
