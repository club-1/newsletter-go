package main

import (
	"log"
	"os"
)

const Name = "newsletterctl"

var l = log.New(os.Stderr, Name+": ", 0)

func main() {
	l.Println("hello world!")
}
