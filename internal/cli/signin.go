package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
)

type signInCommand struct {
	fs   *flag.FlagSet
	name string
}

func SignInCommand() Command {
	cmd := &signInCommand{
		fs: flag.NewFlagSet("login", flag.ExitOnError),
	}

	cmd.fs.StringVar(&cmd.name, "username", "", "Username to sign is as")

	return cmd
}

func (c *signInCommand) Init(args []string) error {
	return c.fs.Parse(args)
}

func (c *signInCommand) Run() {
	reader := bufio.NewReader(os.Stdin)
	if c.name == "" {
		fmt.Print("Enter username: ")
		text, _ := reader.ReadString('\n')
		c.name = text
	}

	fmt.Print("Enter password: ")
	password, _ := reader.ReadString('\n')

	err := signIn(c.name, password)
	if err != nil {
		fmt.Printf("Signing in failed! Error: %v\n", err)
	}
}

func (c *signInCommand) Name() string {
	return c.fs.Name()
}

func (c *signInCommand) Description() string {
	return "Sign in to use the cli"
}

func signIn(name, pw string) error {
	body := map[string]string{
		"username": name,
		"password": pw,
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/auth/sign-in", baseUrl), bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	var resBody map[string]any
	if err = json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("Status %d, and message: %v", res.StatusCode, resBody["message"])
	}

	fmt.Println("Sign in successful!")
	fmt.Printf("To use the cli run the following command:\n$ export OBSERVE_CLI_SESSION=%s\n", resBody["sessionId"])

	return nil
}
