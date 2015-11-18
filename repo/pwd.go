package repo

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/chzyer/readline"
	"github.com/chzyer/readline/runes"

	"github.com/disorganizer/brig/util"
	zxcvbn "github.com/nbutton23/zxcvbn-go"
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

func checkPasswordInBackground(rl *readline.Instance, ticker *time.Ticker) {
	lastPassword := []rune(nil)
	for _ = range ticker.C {
		symbol, color := "", util.Red
		password := rl.Operation.Buf.Runes()

		// Can skip check, input did not change.
		// (save some cpu time...)
		if lastPassword != nil && runes.Equal(password, lastPassword) {
			continue
		}

		strength := zxcvbn.PasswordStrength(string(password), nil)

		switch {
		case strength.Score <= 1:
			symbol = "✗"
			color = util.Red
		case strength.Score <= 2:
			symbol = "⚡"
			color = util.Magenta
		case strength.Score <= 3:
			symbol = "⚠"
			color = util.Yellow
		case strength.Score <= 4:
			symbol = "🗸"
			color = util.Green
		}

		prompt := util.Colorize(symbol, color)
		if strength.Entropy > 0 {
			entropy := fmt.Sprintf(" %3.0f", strength.Entropy)
			prompt += util.Colorize(entropy, util.Cyan)
		} else {
			prompt += util.Colorize(" ENT", util.Cyan)
		}

		prompt += util.Colorize(" New Password: ", color)

		rl.SetPrompt(prompt)
		rl.Operation.Buf.Refresh(nil)
		lastPassword = password
	}
}

// PromptNewPassword asks the user to input a password.
//
// While typing, the user gets feedback by the prompt color,
// which changes with the security of the password to green.
// Additionally the entrtopy of the password is shown.
// If minEntropy was not reached after hitting enter,
// this function will log a message and ask the user again.
func PromptNewPassword(minEntropy float64) (string, error) {
	rl, err := readline.New("")
	if err != nil {
		return "", err
	}
	defer rl.Close()

	// Launch security check in backgroumd
	ticker := time.NewTicker(50 * time.Millisecond)
	go checkPasswordInBackground(rl, ticker)
	defer ticker.Stop()

	for {
		password, err := doPromptLine(rl, "", false)
		if err != nil {
			return "", err
		}

		strength := zxcvbn.PasswordStrength(string(password), nil)
		if strength.Entropy >= minEntropy {
			fmt.Println("\r")
			return password, nil
		}

		log.Println("Password is too weak.")
		log.Printf("Entropy count should reach at least `%3.0f`.\n", minEntropy)
		log.Println("Please try again.")
	}
}

func promptPasswordColored(color int) (string, error) {
	prompt := "Password: "
	if color > 0 {
		prompt = util.Colorize(prompt, color)
	}

	rl, err := readline.New(prompt)
	if err != nil {
		return "", err
	}
	defer rl.Close()

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
	return fmt.Sprintf("Maximum number of password tries exceeded: %d", e.Tries)
}

var triesToColor = map[int]int{
	0: util.Green,
	1: util.Yellow,
	2: util.Magenta,
	3: util.Red,
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
