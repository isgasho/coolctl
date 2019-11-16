package driver

import (
	"encoding/hex"
	"image/color"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
)

// SpeedProfile represents a collection of speed profiles
type SpeedProfile [][]int

func colorFromHexString(c string) (*color.RGBA, error) {
	b, err := hex.DecodeString(c)
	if err != nil {
		return nil, err
	}

	return &color.RGBA{b[0], b[1], b[2], 1}, nil
}

func paletteFromColors(colors []string) (*color.Palette, error) {
	var palette color.Palette
	if colors != nil {
		for _, c := range colors {
			colorCode, err := colorFromHexString(c)
			if err != nil {
				return nil, err
			}
			palette = append(palette, colorCode)
		}
	}

	return &palette, nil
}

func generateSteps(colors color.Palette, mincolors, maxcolors int, mode string, ringonly int) []color.Palette {
	if len(colors) < mincolors {
		log.Fatalf("not enough colors for mode %s, at least %d required", mode, mincolors)
	} else if maxcolors == 0 {
		if len(colors) > 0 {
			log.Printf("too many colors for mode %s, none needed", mode)
			colors = color.Palette{color.RGBA{0, 0, 0, 1}} // discard the input but ensure at least one step
		}
	} else if len(colors) > maxcolors {
		log.Printf("too many colors for mode %s, dropping to %d", mode, maxcolors)
		colors = colors[:maxcolors]
	}

	if len(colors) == 0 {
		colors = color.Palette{color.RGBA{0, 0, 0, 1}}
	}

	var steps []color.Palette

	if !strings.Contains(mode, "super") {
		for colorNum := range colors {
			var colorPalette color.Palette
			for i := 0; i < totalLEDs; i++ {
				colorPalette = append(colorPalette, colors[colorNum])
			}
			steps = append(steps, colorPalette)
		}
	} else if ringonly == 1 {
		steps = append(steps, color.Palette{color.RGBA{0, 0, 0, 1}})
		steps = append(steps, colors)
	} else {
		steps = append(steps, colors)
	}

	return steps
}

func makeRange(min, max, steps int) []int {
	a := make([]int, (max-min)/steps)
	for i := range a {
		a[i] = min
		min += steps
	}

	return a
}

func parseProfile(s string) SpeedProfile {
	d, profiles := SpeedProfile{}, strings.Split(s, "  ")

	for _, profile := range profiles {
		p := strings.Split(profile, " ")
		temp, _ := strconv.Atoi(p[0])
		duty, _ := strconv.Atoi(p[1])
		d = append(d, []int{temp, duty})
	}

	return d
}

func normalizeProfile(p SpeedProfile, temp int) SpeedProfile {
	sort.Slice(p, func(i, j int) bool {
		return p[i][0] < p[j][0]
	})

	lastProfile := p[len(p)-1]
	if lastProfile[0] < temp || lastProfile[1] != 100 {
		if lastProfile[0] == temp || lastProfile[1] == 100 {
			p = p[:len(p)-1]
		}

		p = append(p, []int{temp, 100})
	}

	return p
}

func interpolateProfile(p SpeedProfile) SpeedProfile {
	newProfile, duty, lower, upper := SpeedProfile{}, 0, p[0], p[len(p)-1]

	for _, stdtemp := range makeRange(20, 62, 2) {
		for _, profile := range p {
			if profile[0] <= stdtemp {
				lower = profile
			} else if profile[0] >= stdtemp {
				upper = profile
				break
			}
		}

		if lower[0] == upper[0] {
			duty = lower[1]
		} else {
			duty = int(math.Round(float64(lower[1]) + (float64(stdtemp)-float64(lower[0]))/(float64(upper[0])-float64(lower[0]))*(float64(upper[1])-float64(lower[1]))))
		}

		newProfile = append(newProfile, []int{stdtemp, duty})
	}

	return newProfile
}
