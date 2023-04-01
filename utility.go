package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"

	"github.com/fogleman/gg"
)

type PokedexEntry struct {
	Nat string `json:"Nat"`
}

type APIResult struct {
	Label       string  `json:"label"`
	Probability float64 `json:"probability"`
}

var pokedex map[string][]*PokedexEntry

func init() {
	file, err := os.Open("pokedex.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&pokedex)
	if err != nil {
		log.Fatal(err)
	}
}

func generateImage(r []*APIResult) io.Reader {
	// Load sprites first (located in sprites/)
	sprites := make(map[string]image.Image)
	for _, result := range r {
		sprites[result.Label], _ = gg.LoadPNG("sprites/" + pokedex[result.Label][0].Nat + ".png")
	}

	// Sprites are 128x128 pixels, we space them out by 20 pixels each, and we need a 10 pixel margin on each side, for 3 sprites, that's 10 + 128 + 20 + 128 + 20 + 128 + 10 = 426 pixels
	// For the height, we want to write text above and below the sprites, so we need 2 lines of text, each 20 pixels high, and 10 pixels of margin on each side, for 2 lines of text, that's 10 + 20 + 10 + 20 + 10 = 70 pixels
	// So 128 + 70 = 198 pixels
	dc := gg.NewContext(426, 198)

	// Draw the sprites
	for i, result := range r {
		// Draw the strings
		dc.SetRGB255(255, 255, 255)
		confidence := fmt.Sprintf("%.2f%%", result.Probability)
		dc.DrawStringAnchored(result.Label, float64(10+i*148+64), 138+10, 0.5, 0.5)
		dc.DrawStringAnchored(confidence, float64(10+i*148+64), 138+30, 0.5, 0.5)
		if sprites[result.Label] == nil {
			continue
		}
		dc.DrawImage(sprites[result.Label], 10+i*148, 10)
	}

	dc.SetRGB(0, 0, 0)
	dc.Fill()

	img := dc.Image()

	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	if err != nil {
		log.Fatal(err)
	}

	return buf
}
