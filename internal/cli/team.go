package cli

import (
	"ObservabilityServer/internal/model"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
)

type teamCommand struct {
	fs   *flag.FlagSet
	list bool
}

func TeamCommand() Command {
	cmd := &teamCommand{
		fs: flag.NewFlagSet("teams", flag.ExitOnError),
	}

	cmd.fs.BoolVar(&cmd.list, "list", false, "List teams available to you")

	return cmd
}

func (c *teamCommand) Init(args []string) error {
	return c.fs.Parse(args)
}
func (c *teamCommand) Run() {
	if !c.list {
		fmt.Println("Missing flag")
		c.fs.Usage()
		return
	}

	teams, err := getTeams()
	if err != nil {
		fmt.Printf("Could not get teams! Error: %v\n", err)
		return
	}
	if teams == nil {
		fmt.Printf("No teams were returned!\n")
		return
	}

	fmt.Println("Available teams:")
	fmt.Printf("\tID\tNAME\n")
	for _, team := range teams {
		fmt.Printf("\t%d\t%s\n", team.Id, team.Name)
	}
}
func (c *teamCommand) Name() string {
	return c.fs.Name()
}
func (c *teamCommand) Description() string {
	return "List all your available teams"
}

func getTeams() ([]model.GetTeamDTO, error) {
	secret := os.Getenv("OBSERVE_CLI_SESSION")

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/app/v1/teams", baseUrl), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", secret))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Getting teams failed with status %d\n", res.StatusCode)
	}

	var resBody struct {
		Message string             `json:"message"`
		Teams   []model.GetTeamDTO `json:"teams"`
	}
	if err = json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return nil, err
	}

	return resBody.Teams, nil
}
