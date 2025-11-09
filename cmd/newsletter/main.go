package main

import (
	"fmt"
	"github.com/club-1/newsletter-go"
)

func main() {
	mail := &newsletter.Mail{
		Body: "coucou",
	}
	fmt.Println("newsletter: hello world:", mail)
}
