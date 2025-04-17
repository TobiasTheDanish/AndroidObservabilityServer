package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
)

type apiKeyCommand struct {
	fs     *flag.FlagSet
	create bool
	appId  int
}

func ApiKeyCommand() Command {
	cmd := &apiKeyCommand{
		fs: flag.NewFlagSet("keys", flag.ExitOnError),
	}

	cmd.fs.BoolVar(&cmd.create, "create", false, "Create a new api key for app id specified with 'id'")
	cmd.fs.IntVar(&cmd.appId, "id", -1, "Id of the app the api key belongs to")

	return cmd
}

func (c *apiKeyCommand) Init(args []string) error {
	return c.fs.Parse(args)
}
func (c *apiKeyCommand) Run() {
	if !c.create || c.appId == -1 {
		fmt.Println("Malformed arguments for 'keys' command")
		c.fs.Usage()
	} else {
		err := createApiKey(c.appId)
		if err != nil {
			fmt.Printf("Could not create apiKey: %v\n", err)
		}
	}
}
func (c *apiKeyCommand) Name() string {
	return c.fs.Name()
}
func (c *apiKeyCommand) Description() string {
	return "Generate new api keys for owners"
}

func createApiKey(appId int) error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/auth/apps/%d/keys", baseUrl, appId), nil)
	if err != nil {
		return err
	}
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
		return fmt.Errorf("Creating new api key for app id '%d' failed with status %d, and body: \n%v\nMAKE SURE ENV VAR 'CLI_SECRET' IS SET!", appId, res.StatusCode, resBody)
	}

	fmt.Printf("New api key created!\nReturned key:\n\t%v\n", resBody["key"])

	return nil

}
