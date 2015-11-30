package repo

import (
	"fmt"
	"io"
	"math"
	"os/exec"
	"time"
)

type EffectQueue struct {
	StdInPipe io.Writer
}

func (q *EffectQueue) Push(e Effect) {
	c := e.ComposeEffect()
	for color := range c {
		colorValue := fmt.Sprintf("%d %d %d\n", color.R, color.G, color.B)
		q.StdInPipe.Write([]byte(colorValue))
	}
}

func CreateEffectQueue() (*EffectQueue, error) {
	cmd := exec.Command("catlight", "cat")
	stdinpipe, err := cmd.StdinPipe()
	if err != nil {
		return &EffectQueue{StdInPipe: nil}, err
	}
	return &EffectQueue{stdinpipe}, cmd.Start()
}

type SimpleColor struct {
	R, G, B uint8
}

type Effect interface {
	ComposeEffect() chan SimpleColor
}

type Properties struct {
	Delay  time.Duration
	Color  SimpleColor
	Repeat int
	Cancel bool
}

type FadeEffect struct {
	Properties
}

type FlashEffect struct {
	Properties
}

func (color *SimpleColor) ComposeEffect() chan SimpleColor {
	c := make(chan SimpleColor, 1)
	c <- SimpleColor{color.R, color.G, color.B}
	close(c)
	return c
}

func (effect *FlashEffect) ComposeEffect() chan SimpleColor {
	c := make(chan SimpleColor, 1)
	keepLooping := false
	if effect.Repeat <= 0 {
		keepLooping = true
	}
	go func() {
		for {
			if effect.Repeat <= 0 && !keepLooping {
				break
			}
			c <- effect.Color
			time.Sleep(effect.Delay)
			c <- SimpleColor{0, 0, 0}
			time.Sleep(effect.Delay)
			effect.Repeat--
		}
		close(c)
	}()
	return c
}

func max(r uint8, g uint8, b uint8) uint8 {
	max := r
	if max < g {
		max = g
	}
	if max < b {
		max = b
	}
	return max
}

func (effect *FadeEffect) ComposeEffect() chan SimpleColor {
	c := make(chan SimpleColor, 1)

	keepLooping := false
	if effect.Repeat <= 0 {
		keepLooping = true
	}

	max := max(effect.Color.R, effect.Color.B, effect.Color.G)
	go func() {
		for {

			if effect.Repeat <= 0 && !keepLooping {
				break
			}

			r := int(math.Floor(float64(effect.Color.R) / float64(max) * 100.0))
			g := int(math.Floor(float64(effect.Color.G) / float64(max) * 100.0))
			b := int(math.Floor(float64(effect.Color.B) / float64(max) * 100.0))

			for i := 0; i < int(max); i += 1 {
				c <- SimpleColor{uint8((i * r) / 100), uint8((i * g) / 100), uint8((i * b) / 100)}
				time.Sleep(effect.Delay)
			}

			for i := int(max - 1); i >= 0; i -= 1 {
				c <- SimpleColor{uint8((i * r) / 100), uint8((i * g) / 100), uint8((i * b) / 100)}
				time.Sleep(effect.Delay)
			}
			effect.Repeat--
		}
		close(c)
	}()
	return c
}

type BlendEffect struct {
	StartColor SimpleColor
	EndColor   SimpleColor
	Duration   time.Duration
}

func (effect *BlendEffect) ComposeEffect() chan SimpleColor {
	c := make(chan SimpleColor, 1)
	go func() {
		// How much colors should be generated during the effect?
		N := 20 * effect.Duration.Seconds()

		sr := float64(effect.StartColor.R)
		sg := float64(effect.StartColor.G)
		sb := float64(effect.StartColor.B)

		for i := 0; i < int(N); i++ {
			sr += (float64(effect.EndColor.R) - float64(effect.StartColor.R)) / N
			sg += (float64(effect.EndColor.G) - float64(effect.StartColor.G)) / N
			sb += (float64(effect.EndColor.B) - float64(effect.StartColor.B)) / N

			c <- SimpleColor{uint8(sr), uint8(sg), uint8(sb)}
			time.Sleep(time.Duration(1/N*1000) * time.Millisecond)
		}

		close(c)
	}()

	return c
}
