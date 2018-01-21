package pwd

import (
	"bytes"
	"fmt"

	"github.com/chzyer/readline"
	"github.com/fatih/color"

	zxcvbn "github.com/nbutton23/zxcvbn-go"
	"github.com/sahib/brig/util"
)

const (
	msgLowEntropy  = "\nPlease enter a password with at least %g bits entropy."
	msgReEnter     = "\nWell done! Please re-type your password now:"
	msgBadPassword = "\nThis did not seem to match. Please retype it again."
	msgMaxTriesHit = "\nMaximum number of password tries exceeded: %d"
)

func doPromptLine(rl *readline.Instance, prompt string, hide bool) (string, error) {
	var line = ""
	var bytepwd []byte
	var err error

	if hide {
		bytepwd, err = rl.ReadPassword(prompt)
		line = string(bytepwd)
	} else {
		line, err = rl.Readline()
	}

	if err != nil {
		return "", err
	}

	return line, nil
}

func createStrengthPrompt(password []rune, prefix string) string {
	symbol, colorFn := "", color.RedString
	strength := zxcvbn.PasswordStrength(string(password), nil)

	switch {
	case strength.Score <= 1:
		symbol = "✗"
		colorFn = color.RedString
	case strength.Score <= 2:
		symbol = "⚡"
		colorFn = color.MagentaString
	case strength.Score <= 3:
		symbol = "⚠"
		colorFn = color.YellowString
	case strength.Score <= 4:
		symbol = "✔"
		colorFn = color.GreenString
	}

	prompt := colorFn(symbol)
	if strength.Entropy > 0 {
		entropy := fmt.Sprintf(" %3.0f", strength.Entropy)
		prompt += color.CyanString(entropy)
	} else {
		prompt += color.CyanString("   0")
	}

	prompt += colorFn(" " + prefix + "passphrase: ")
	return prompt
}

// PromptNewPassword asks the user to input a password.
//
// While typing, the user gets feedback by the prompt color,
// which changes with the security of the password to green.
// Additionally the entrtopy of the password is shown.
// If minEntropy was not reached after hitting enter,
// this function will log a message and ask the user again.
func PromptNewPassword(minEntropy float64) ([]byte, error) {
	rl, err := readline.New("")
	if err != nil {
		return nil, err
	}
	defer util.Closer(rl)

	passwordCfg := rl.GenPasswordConfig()
	passwordCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		rl.SetPrompt(createStrengthPrompt(line, "New "))
		rl.Refresh()
		return nil, 0, false
	})

	pwd := []byte{}

	for {
		pwd, err = rl.ReadPasswordWithConfig(passwordCfg)
		if err != nil {
			return nil, err
		}

		strength := zxcvbn.PasswordStrength(string(pwd), nil)
		if strength.Entropy >= minEntropy {
			break
		}

		fmt.Printf(color.YellowString(msgLowEntropy)+"\n", minEntropy)
	}

	passwordCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		rl.SetPrompt(createStrengthPrompt(line, "Retype "))
		rl.Refresh()
		return nil, 0, false
	})

	fmt.Println(color.GreenString(msgReEnter))

	for {
		newPwd, err := rl.ReadPasswordWithConfig(passwordCfg)
		if err != nil {
			return nil, err
		}

		if bytes.Equal(pwd, newPwd) {
			break
		}

		fmt.Println(color.YellowString(msgBadPassword))
	}

	return pwd, nil
}

func promptPassword(prompt string) (string, error) {
	rl, err := readline.New(prompt)
	if err != nil {
		return "", err
	}
	defer util.Closer(rl)

	return doPromptLine(rl, prompt, true)
}

// PromptPassword just opens an uncolored password prompt.
//
// The password is not echo'd to stdout for safety reasons.
func PromptPassword() (string, error) {
	return promptPassword("Password: ")
}

// ErrTooManyTries happens when the user failed the password check too often
type ErrTooManyTries struct {
	Tries int
}

func (e ErrTooManyTries) Error() string {
	return fmt.Sprintf(msgMaxTriesHit, e.Tries)
}

var triesToColor = map[int]func(string, ...interface{}) string{
	0: color.GreenString,
	1: color.YellowString,
	2: color.MagentaString,
	3: color.RedString,
}

// PromptPasswordMaxTries tries to read a password maxTries times.
//
// The typed password can be validated by the caller via the passfn function.
// If the user failed to pass the correct password, ErrTooManyTries is returned.
// For visual guidance the prompt color will gradually change from green to red
// with each failed try.
func PromptPasswordMaxTries(maxTries int, passfn func(string) bool) (string, error) {
	for i := 0; i < maxTries; i++ {
		colorFn := triesToColor[util.Min(i, len(triesToColor))]
		pwd, err := promptPassword(colorFn("Password: "))
		if err != nil {
			return "", err
		}

		if !passfn(pwd) {
			continue
		}

		return pwd, err
	}

	return "", ErrTooManyTries{maxTries}
}
