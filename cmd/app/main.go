package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/seshoo/bookFinder/internal/commands"
)

var opts struct {
	ParseCmd commands.ParseCommand `command:"parse"`
	BotCmd   commands.BotCommand   `command:"bot"`
}

var revision = "undefined"

var exitFunc = os.Exit

func main() {
	fmt.Printf("Book finder: %s\n", revision)

	p := flags.NewParser(&opts, flags.Default)
	if _, err := p.Parse(); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && errors.Is(flagsErr.Type, flags.ErrHelp) {
			exitFunc(0)
		}
		exitFunc(1)
	}
}
