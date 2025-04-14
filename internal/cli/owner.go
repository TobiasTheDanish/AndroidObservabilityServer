package cli

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
)

type ownerCommand struct {
	fs     *flag.FlagSet
	create bool
	name   string
	list   bool
}

func OwnerCommand() Command {
	cmd := &ownerCommand{
		fs: flag.NewFlagSet("owner", flag.ExitOnError),
	}

	cmd.fs.BoolVar(&cmd.create, "create", false, "Create a new owner. Name must be specified when creating a new owner")
	cmd.fs.StringVar(&cmd.name, "name", "", "Name of the owner")
	cmd.fs.BoolVar(&cmd.list, "list", false, "List all current owners. Not implmented yet")

	return cmd
}

func (c *ownerCommand) Init(args []string) error {
	return c.fs.Parse(args)
}

// TODO: add support for list command
func (c *ownerCommand) Run() {

	if !(c.create || c.list) {
		fmt.Println("Missing create or list flag")
		c.fs.Usage()
		return
	}

	if c.create && c.list {
		fmt.Println("Cannot both create new and list current owners at the same time")
		c.fs.Usage()
		return
	}

	if c.create && c.name == "" {
		fmt.Println("Name must be specified when creating owner")
		c.fs.Usage()
		return
	}

	if c.create {
		err := createOwner(c.name)
		if err != nil {
			fmt.Printf("Error creating new owner: %v\n", err)
		}
		return
	}

	c.fs.Usage()
	return
}

func (c *ownerCommand) Name() string {
	return c.fs.Name()
}

func (c *ownerCommand) Description() string {
	return "Create new or list current owners"
}

func createOwner(name string) error {
	body := map[string]string{
		"name": name,
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/auth/owners", baseUrl), bytes.NewReader(jsonBytes))
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
		return fmt.Errorf("Creating new owner failed with status %d, and body: \n%v\nMAKE SURE ENV VAR 'CLI_SECRET' IS SET!", res.StatusCode, resBody)
	}

	fmt.Printf("New owner created!\nReturned id: '%v'\n", resBody["id"])

	return nil
}
