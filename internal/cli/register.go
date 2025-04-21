package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type registerCommand struct {
	fs       *flag.FlagSet
	username string
	teamname string
}

func RegisterCommand() Command {
	cmd := &registerCommand{
		fs: flag.NewFlagSet("register", flag.ExitOnError),
	}

	cmd.fs.StringVar(&cmd.username, "user", "", "Your username, used to login")
	cmd.fs.StringVar(&cmd.teamname, "team", "", "Name of your first team")

	return cmd
}

func (c *registerCommand) Init(args []string) error {
	return c.fs.Parse(args)
}

func (c *registerCommand) Run() {
	reader := bufio.NewReader(os.Stdin)
	if c.username == "" {
		fmt.Print("Enter username: ")
		text, _ := reader.ReadString('\n')
		c.username = strings.TrimSpace(text)
	}

	fmt.Print("Enter password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if c.teamname == "" {
		fmt.Print("Enter your team name: ")
		text, _ := reader.ReadString('\n')
		c.teamname = strings.TrimSpace(text)
	}

	fmt.Println("Registering user")
	userId, err := register(c.username, password)
	if err != nil {
		fmt.Printf("Registration failed! Error: %v\n", err)
		return
	}

	fmt.Println("Signing in")
	sessionId, err := signInNewUser(c.username, password)
	if err != nil {
		fmt.Printf("Sign in failed! Error: %v\n", err)
		return
	}

	// Create team from teamname, get team id
	fmt.Println("Creating team")
	teamId, err := createTeam(c.teamname, sessionId)
	if err != nil {
		fmt.Printf("Team creation failed! Error: %v\n", err)
		return
	}

	// Create team-user link
	fmt.Println("Creating team-user link")
	err = createTeamUserLink(teamId, userId, sessionId)
	if err != nil {
		fmt.Printf("Team-User link creation failed! Error: %v\n", err)
		return
	}

	fmt.Printf("Successfully created a new user and team!\n    UserID = %d\n    TeamID = %d\n", userId, teamId)

	fmt.Printf("To use the cli run the following command:\n$ export OBSERVE_CLI_SESSION=%s\n", sessionId)
}

func (c *registerCommand) Name() string {
	return c.fs.Name()
}

func (c *registerCommand) Description() string {
	return "Register to use the cli"
}

func register(username, password string) (int, error) {
	body := map[string]string{
		"name":     username,
		"password": password,
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return -1, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/auth/register", baseUrl), bytes.NewReader(jsonBytes))
	if err != nil {
		return -1, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}

	var resBody map[string]any
	if err = json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return -1, err
	}

	if res.StatusCode != http.StatusCreated {
		return -1, fmt.Errorf("Status %d, and message: %v", res.StatusCode, resBody["message"])
	}

	fmt.Println("Registration successful!")

	userId := resBody["id"].(float64)

	return int(userId), nil
}

func signInNewUser(name, pw string) (string, error) {
	body := map[string]string{
		"username": name,
		"password": pw,
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/auth/sign-in", baseUrl), bytes.NewReader(jsonBytes))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	var resBody map[string]any
	if err = json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("Status %d, and message: %v", res.StatusCode, resBody["message"])
	}

	fmt.Println("Sign in successful!")

	return resBody["sessionId"].(string), nil
}

func createTeam(teamname, sessionId string) (int, error) {
	body := map[string]string{
		"name": teamname,
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return -1, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/app/v1/teams", baseUrl), bytes.NewReader(jsonBytes))
	if err != nil {
		return -1, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sessionId))
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}

	var resBody map[string]any
	if err = json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return -1, err
	}

	if res.StatusCode != http.StatusCreated {
		return -1, fmt.Errorf("Status %d, and message: %v", res.StatusCode, resBody["message"])
	}

	fmt.Println("Team created successfully!")

	userId := resBody["id"].(float64)

	return int(userId), nil
}

func createTeamUserLink(teamId, userId int, sessionId string) error {
	body := map[string]any{
		"userId": userId,
		"role":   "owner",
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/app/v1/teams/%d/users", baseUrl, teamId), bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sessionId))
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

	fmt.Println("Team-User link created successfully!")

	return nil
}
