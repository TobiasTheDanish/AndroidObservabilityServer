package cli

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
)

type appCommand struct {
	fs     *flag.FlagSet
	create bool
	name   string
	teamId int
	list   bool
}

func AppCommand() Command {
	cmd := &appCommand{
		fs: flag.NewFlagSet("app", flag.ExitOnError),
	}

	cmd.fs.BoolVar(&cmd.create, "create", false, "Create a new app. Name must be specified when creating a new app")
	cmd.fs.StringVar(&cmd.name, "name", "", "Name of the app")
	cmd.fs.IntVar(&cmd.teamId, "teamId", -1, "Id of the team, which the app belongs to")
	cmd.fs.BoolVar(&cmd.list, "list", false, "List all current apps. Not implmented yet")

	return cmd
}

func (c *appCommand) Init(args []string) error {
	return c.fs.Parse(args)
}

// TODO: add support for list command
func (c *appCommand) Run() {
	if !(c.create || c.list) {
		fmt.Println("Missing create or list flag")
		c.fs.Usage()
		return
	}

	if c.create && c.list {
		fmt.Println("Cannot both create new and list current apps at the same time")
		c.fs.Usage()
		return
	}

	if c.create && (c.name == "" || c.teamId == -1) {
		fmt.Println("Name and teamId must be specified when creating app")
		c.fs.Usage()
		return
	}

	if c.create {
		err := createApp(c.name, c.teamId)
		if err != nil {
			fmt.Printf("Error creating new app: %v\n", err)
		}
		return
	}

	c.fs.Usage()
	return
}

func (c *appCommand) Name() string {
	return c.fs.Name()
}

func (c *appCommand) Description() string {
	return "Create new or list current apps"
}

func createApp(name string, teamId int) error {
	secret := os.Getenv("OBSERVE_CLI_SESSION")
	body := map[string]any{
		"name":   name,
		"teamId": teamId,
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/app/v1/apps", baseUrl), bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", secret))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	var resBody map[string]any
	if err = json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("Creating new app failed with status %d, and body: \n%v\nMAKE SURE ENV VAR 'CLI_SECRET' IS SET!", res.StatusCode, resBody)
	}

	fmt.Printf("New app created!\nReturned id: '%v'\n", resBody["id"])

	return nil
}
