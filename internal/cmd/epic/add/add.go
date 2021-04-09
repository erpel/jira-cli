package add

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/internal/query"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	helpText = `Add issues to an epic.`
	examples = `$ jira epic add EPIC_KEY ISSUE_1 ISSUE_2`
)

// NewCmdAdd is an add command.
func NewCmdAdd() *cobra.Command {
	return &cobra.Command{
		Use:     "add EPIC_KEY ISSUE_1 [...ISSUE_N]",
		Short:   "Add issues to an epic",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"assign"},
		Annotations: map[string]string{
			"help:args": "EPIC_KEY\t\tEpic to which you want to assign issues to, eg: EPIC-1\n" +
				"ISSUE_1 [...ISSUE_N]\tKey of the issues to add to an epic (max 50 issues at once)",
		},
		Run: add,
	}
}

func add(cmd *cobra.Command, args []string) {
	server := viper.GetString("server")
	params := parseFlags(cmd.Flags(), args)
	client := api.Client(jira.Config{Debug: params.debug})

	qs := getQuestions(params)
	if len(qs) > 0 {
		ans := struct {
			EpicKey string
			Issues  string
		}{}
		err := survey.Ask(qs, &ans)
		cmdutil.ExitIfError(err)

		if params.epicKey == "" {
			params.epicKey = ans.EpicKey
		}

		if len(params.issues) == 0 {
			issues := strings.Split(ans.Issues, ",")
			for i, iss := range issues {
				issues[i] = strings.TrimSpace(iss)
			}
			params.issues = issues
		}
	}

	err := func() error {
		s := cmdutil.Info("Adding issues to the epic...")
		defer s.Stop()

		return client.EpicIssuesAdd(params.epicKey, params.issues...)
	}()
	cmdutil.ExitIfError(err)

	fmt.Printf("\033[0;32m✓\033[0m Issues added to the epic %s\n%s/browse/%s\n", params.epicKey, server, params.epicKey)
}

func parseFlags(flags query.FlagParser, args []string) *addParams {
	var (
		epicKey string
		issues  []string
	)

	nArgs := len(args)
	if nArgs > 0 {
		epicKey = args[0]
	}
	if nArgs > 1 {
		issues = args[1:]
	}

	debug, err := flags.GetBool("debug")
	cmdutil.ExitIfError(err)

	return &addParams{
		epicKey: epicKey,
		issues:  issues,
		debug:   debug,
	}
}

func getQuestions(params *addParams) []*survey.Question {
	var qs []*survey.Question

	if params.epicKey == "" {
		qs = append(qs, &survey.Question{
			Name:     "epicKey",
			Prompt:   &survey.Input{Message: "Epic Key"},
			Validate: survey.Required,
		})
	}
	if len(params.issues) == 0 {
		qs = append(qs, &survey.Question{
			Name: "issues",
			Prompt: &survey.Input{
				Message: "Issues",
				Help:    "Comma separated list of issues key to add. eg: ISSUE_1, ISSUE_2",
			},
			Validate: survey.Required,
		})
	}

	return qs
}

type addParams struct {
	epicKey string
	issues  []string
	debug   bool
}