package commands

import "fmt"

type BotCommand struct {
	CommonOpts
}

func (c *BotCommand) Execute(_ []string) error {
	fmt.Println("Bot Command")
	fmt.Printf("Debug Mode: %t\n", c.Dbg)

	// todo
	return nil
}
