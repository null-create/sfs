package client

import (
	"fmt"
	"log"
	"strings"
)

// prompt the user whether they want to continue
func (c *Client) Continue() bool {
	var ans string
	fmt.Print("continue? (y/n): ")
	fmt.Scanln(&ans)
	return strings.ToLower(ans) == "y"
}

func (c *Client) getInput(prompt, input string) {
	fmt.Print(prompt)
	_, err := fmt.Scanln(&input)
	if err != nil {
		log.Fatal(err)
	}
	if input == "" {
		log.Fatal("input not found")
	}
}
