package cli

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	baseUrl = os.Getenv("CLI_BASE_URL")
)

type Command interface {
	Init(args []string) error
	Run()
	Name() string
	Description() string
}

type Commands []Command

func (c Commands) Usage() {
	fmt.Println("Available subcommands:")
	for _, cmd := range c {
		fmt.Printf("\t%s\t%s\n", cmd.Name(), cmd.Description())
	}
}

func ParseFlags(args []string) (bool, Command) {
	cmds := Commands{
		AppCommand(),
		ApiKeyCommand(),
		RegisterCommand(),
		SignInCommand(),
	}

	if len(args) < 1 {
		fmt.Println("You must pass a subcommand")
		cmds.Usage()
		return false, nil
	}

	subCmd := args[0]

	if subCmd == "help" {
		cmds.Usage()
		return false, nil
	}

	for _, cmd := range cmds {
		if cmd.Name() == subCmd {
			err := cmd.Init(args[1:])
			if err != nil {
				fmt.Println(err.Error())
				flag.Usage()
				return false, nil
			}

			return true, cmd
		}
	}

	fmt.Printf("Invalid subcommand %s\n", subCmd)
	flag.Usage()
	return false, nil
}
