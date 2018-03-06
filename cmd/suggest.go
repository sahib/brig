package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/urfave/cli"
	"github.com/xrash/smetrics"
)

type suggestion struct {
	name  string
	score float64
}

func levenshteinRatio(s, t string) float64 {
	lensum := float64(len(s) + len(t))
	if lensum == 0 {
		return 1.0
	}

	dist := float64(smetrics.WagnerFischer(s, t, 1, 1, 2))
	return (lensum - dist) / lensum
}

func findLastGoodCommands(ctx *cli.Context) ([]string, []cli.Command) {
	for ctx.Parent() != nil {
		ctx = ctx.Parent()
	}

	args := ctx.Args()
	if len(args) == 0 || len(args) == 1 {
		return nil, ctx.App.Commands
	}

	cmd := ctx.App.Command(args[0])
	if cmd == nil {
		return nil, ctx.App.Commands
	}

	validArgs := []string{args[0]}
	args = args[1 : len(args)-1]

	for len(args) != 0 && cmd != nil {
		for _, subCmd := range cmd.Subcommands {
			if subCmd.Name == args[0] {
				cmd = &subCmd
			}
		}

		validArgs = append(validArgs, args[0])
		args = args[1:]
	}

	return validArgs, cmd.Subcommands
}

func findSimilarCommands(cmdName string, cmds []cli.Command) []suggestion {
	similars := []suggestion{}

	for _, cmd := range cmds {
		candidates := []string{cmd.Name}
		candidates = append(candidates, cmd.Aliases...)

		for _, candidate := range candidates {
			if score := levenshteinRatio(cmdName, candidate); score >= 0.6 {
				similars = append(similars, suggestion{
					name:  cmd.Name,
					score: score,
				})
				break
			}
		}
	}

	// Special cases for the git inclined:
	staticSuggestions := map[string]string{
		"insert": "stage",
		"pull":   "sync",
	}

	for gitName, brigName := range staticSuggestions {
		if cmdName == gitName {
			similars = append(similars, suggestion{
				name:  brigName,
				score: 0.0,
			})
		}
	}

	// Let suggestions be sorted by their similarity:
	sort.Slice(similars, func(i, j int) bool {
		return similars[i].score < similars[j].score
	})

	return similars
}

func commandNotFound(ctx *cli.Context, cmdName string) {
	// Try to find the commands we need to look at for a suggestion/
	// We only want to show the user the relevant subcommands.
	cmdPath, lastGoodCmds := findLastGoodCommands(ctx)

	// Figure out if it was a toplevel command or if some subcommand
	// (like e.g. 'remote') was correct.
	badCmd := color.RedString(cmdName)
	if cmdPath == nil {
		// A toplevel command was wrong:
		fmt.Printf("`%s` is not a valid command. ", badCmd)
	} else {
		// A command of a subcommand was wrong:
		lastGoodSubCmd := color.YellowString(strings.Join(cmdPath, " "))
		fmt.Printf("`%s` is not a valid subcommand of `%s`. ", badCmd, lastGoodSubCmd)
	}

	// Get a list of similar commands:
	similars := findSimilarCommands(cmdName, lastGoodCmds)

	switch len(similars) {
	case 0:
		fmt.Printf("\n")
	case 1:
		suggestion := color.GreenString(similars[0].name)
		fmt.Printf("Did you maybe mean `%s`?\n", suggestion)
	default:
		fmt.Println("\n\nDid you maybe mean one of those?")
		for _, similar := range similars {
			fmt.Printf("  * %s\n", color.GreenString(similar.name))
		}
	}
}
