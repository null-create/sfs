package client

import (
	"fmt"
	"strings"
)

// prompt the user whether they want to continue
func (c *Client) Continue() bool {
	var ans string
	fmt.Print("continue? (y/n): ")
	fmt.Scanln(&ans)
	return strings.ToLower(ans) == "y"
}
