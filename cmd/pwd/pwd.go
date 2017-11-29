package pwd

import (
	"bytes"
	"fmt"

	"github.com/chzyer/readline"

	"github.com/disorganizer/brig/util"
	"github.com/disorganizer/brig/util/colors"
	zxcvbn "github.com/nbutton23/zxcvbn-go"
)

const (
	msgLowEntropy  = "⚠ Please enter a password with at least %g bits entropy."
	msgReEnter     = "✔ Well done! Please re-type your password now:"
	msgBadPassword = "⚠ This did not seem to match. Please try again."
	msgMaxTriesHit = "⚡ Maximum number of password tries exceeded: %d"
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
	symbol, color := "", colors.Red
	strength := zxcvbn.PasswordStrength(string(password), nil)

	switch {
	case strength.Score <= 1:
		symbol = "✗"
		color = colors.Red
	case strength.Score <= 2:
		symbol = "⚡"
		color = colors.Magenta
	case strength.Score <= 3:
		symbol = "⚠"
		color = colors.Yellow
	case strength.Score <= 4:
		symbol = "✔"
		color = colors.Green
	}

	prompt := colors.Colorize(symbol, color)
	if strength.Entropy > 0 {
		entropy := fmt.Sprintf(" %3.0f", strength.Entropy)
		prompt += colors.Colorize(entropy, colors.Cyan)
	} else {
		prompt += colors.Colorize(" ENT", colors.Cyan)
	}

	prompt += colors.Colorize(" "+prefix+"passphrase: ", color)
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
		rl.SetPrompt(createStrengthPrompt(line, "   New "))
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

		fmt.Printf(colors.Colorize(msgLowEntropy, colors.Yellow)+"\n", minEntropy)
	}

	passwordCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		rl.SetPrompt(createStrengthPrompt(line, "Retype "))
		rl.Refresh()
		return nil, 0, false
	})

	fmt.Println(colors.Colorize(msgReEnter, colors.Green))

	for {
		newPwd, err := rl.ReadPasswordWithConfig(passwordCfg)
		if err != nil {
			return nil, err
		}

		if bytes.Equal(pwd, newPwd) {
			break
		}

		fmt.Println(colors.Colorize(msgBadPassword, colors.Yellow))
	}

	return pwd, nil
}

func promptPasswordColored(color int) (string, error) {
	prompt := "Password: "
	if color > 0 {
		prompt = colors.Colorize(prompt, color)
	}

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
	return promptPasswordColored(0)
}

// ErrTooManyTries happens when the user failed the password check too often
type ErrTooManyTries struct {
	Tries int
}

func (e ErrTooManyTries) Error() string {
	return fmt.Sprintf(msgMaxTriesHit, e.Tries)
}

var triesToColor = map[int]int{
	0: colors.Green,
	1: colors.Yellow,
	2: colors.Magenta,
	3: colors.Red,
}

// PromptPasswordMaxTries tries to read a password maxTries times.
//
// The typed password can be validated by the caller via the passfn function.
// If the user failed to pass the correct password, ErrTooManyTries is returned.
// For visual guidance the prompt color will gradually change from green to red
// with each failed try.
func PromptPasswordMaxTries(maxTries int, passfn func(string) bool) (string, error) {
	for i := 0; i < maxTries; i++ {
		color := triesToColor[util.Min(i, len(triesToColor))]
		pwd, err := promptPasswordColored(color)
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
