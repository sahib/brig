package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/disorganizer/brig/util/colors"
	"github.com/urfave/cli"
)

// ld compares two strings and returns the levenshtein distance between them.
// TODO: Use a proper library for that.
func levenshtein(s, t string) float64 {
	s = strings.ToLower(s)
	t = strings.ToLower(t)

	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
	}
	for i := range d {
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for j := 1; j <= len(t); j++ {
		for i := 1; i <= len(s); i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				min := d[i-1][j]
				if d[i][j-1] < min {
					min = d[i][j-1]
				}
				if d[i-1][j-1] < min {
					min = d[i-1][j-1]
				}
				d[i][j] = min + 1
			}
		}

	}

	// Return the levenshtein ratio, rather than the absolute distance:
	total_len := len(s)
	if len(t) > total_len {
		total_len = len(t)
	}

	dist := d[len(s)][len(t)]
	return float64(dist) / float64(total_len)
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

type suggestion struct {
	name  string
	score float64
}

func findSimilarCommands(cmdName string, cmds []cli.Command) []suggestion {
	similars := []suggestion{}

	for _, cmd := range cmds {
		candidates := []string{cmd.Name}
		candidates = append(candidates, cmd.Aliases...)

		for _, candidate := range candidates {
			score := levenshtein(cmdName, candidate)
			if score <= 0.5 {
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
		"add":  "stage",
		"pull": "sync",
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
	badCmd := colors.Colorize(cmdName, colors.Red)
	if cmdPath == nil {
		// A toplevel command was wrong:
		fmt.Printf("`%s` is not a valid command. ", badCmd)
	} else {
		// A command of a subcommand was wrong:
		lastGoodSubCmd := colors.Colorize(strings.Join(cmdPath, " "), colors.Yellow)
		fmt.Printf("`%s` is not a valid subcommand of `%s`. ", badCmd, lastGoodSubCmd)
	}

	// Get a list of similar commands:
	similars := findSimilarCommands(cmdName, lastGoodCmds)

	switch len(similars) {
	case 0:
		fmt.Printf("\n")
	case 1:
		suggestion := colors.Colorize(similars[0].name, colors.Green)
		fmt.Printf("Did you maybe mean `%s`?\n", suggestion)
	default:
		fmt.Println("\n\nDid you mean one of those?")
		for _, similar := range similars {
			fmt.Printf("  * %s\n", colors.Colorize(similar.name, colors.Green))
		}
	}
}
