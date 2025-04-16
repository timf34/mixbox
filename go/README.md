# Mixbox-Go

Below is a sample README.md you could use for your Go implementation. You can adjust the wording to suit your style.

---

```markdown
# Mixbox for Go

This repository provides a Go implementation of the Mixbox pigment mixing model, originally described in the SIGGRAPH 2021 paper ["Practical Pigment Mixing for Digital Painting"](https://doi.org/10.1145/3478513.3480549) by Sochorová and Jamriška.  
Mixbox simulates realistic pigment blending through a latent color space, making pigment mixing more true to real-world behavior than standard RGB linear interpolation.

## Installation

Initialize a Go module in your project (if you haven’t already):

```bash
go mod init github.com/timf34/mixbox-go
```

To use the Mixbox package in your project, simply add:

```bash
go get github.com/timf34/mixbox-go/mixbox
```

## Usage

### Basic Color Mixing

Below is a minimal example that uses the Mixbox library to linearly blend two colors (using pigment-based interpolation):

```go
package main

import (
	"fmt"
	"github.com/timf34/mixbox-go/mixbox"
)

func main() {
	rgb1 := [3]uint8{0, 33, 133}   // e.g. blue
	rgb2 := [3]uint8{252, 211, 0}   // e.g. yellow
	t := 0.5                      // mixing ratio

	// Performs pigment mixing using the latent space interpolation.
	rgbMix := mixbox.Lerp(rgb1, rgb2, t)
	fmt.Printf("Mixed color: %d %d %d\n", rgbMix[0], rgbMix[1], rgbMix[2])
}
```

### Mixing Multiple Colors

You can perform multi-color blending by converting RGB colors to their latent representations, mixing them with weighted contributions, and finally converting back to RGB. For example:

```go
package main

import (
	"fmt"
	"github.com/timf34/mixbox-go/mixbox"
)

func main() {
	// Example colors
	rgb1 := [3]uint8{254, 236, 0}  // Cadmium Yellow
	rgb2 := [3]uint8{255, 39, 2}    // Cadmium Red
	rgb3 := [3]uint8{25, 0, 89}     // Ultramarine Blue

	// Convert each color into the latent representation.
	latent1 := mixbox.RGBToLatent(rgb1)
	latent2 := mixbox.RGBToLatent(rgb2)
	latent3 := mixbox.RGBToLatent(rgb3)

	var latentMix [mixbox.LatentSize]float64

	// Mix with weights 0.3, 0.6, and 0.1 respectively.
	for i := 0; i < mixbox.LatentSize; i++ {
		latentMix[i] = 0.3*latent1[i] + 0.6*latent2[i] + 0.1*latent3[i]
	}

	// Convert the mixed latent representation back to RGB.
	rgbMix := mixbox.LatentToRGB(latentMix)
	fmt.Printf("Multi-color mixed result: %d %d %d\n", rgbMix[0], rgbMix[1], rgbMix[2])
}
```

### Running the Demo: Gradient Comparison

A demo application is provided in `main.go` that generates a PNG image (`gradient.png`). It compares Mixbox blending against standard linear RGB mixing. To run the demo, execute:

```bash
go run main.go
```

The demo loads the pigment lookup table (LUT) from `lut.dat` (or you can use the provided string-based decompression helper) and creates an image where the top half shows Mixbox (pigment-based) blending and the bottom half shows linear RGB interpolation.

## LUT Initialization

The library supports two approaches to initialize the LUT:

1. **From a Binary File:**  
   The demo (in `main.go`) uses:
   ```go
   err := mixbox.LoadLUTFromFile("lut.dat")
   ```
2. **From a Raw Encoded String:**  
   You can also decompress a base64-encoded, deflate-compressed string (as in the Python implementation) via:
   ```go
   err := mixbox.DecompressAndInitLUT("xNrFmuTYEQXgV5mtmZn5AcbMzDRcK")
   ```

Choose the method that best suits your distribution or integration needs.

## Pigment Colors
| Pigment |  | RGB | Float RGB | Linear RGB |
| --- | --- |:----:|:----:|:----:|
| Cadmium Yellow | <img src="https://scrtwpns.com/mixbox/pigments/cadmium_yellow.png"/> | 254, 236, 0  | 0.996, 0.925, 0.0 | 0.991, 0.839, 0.0 |
| Hansa Yellow | <img src="https://scrtwpns.com/mixbox/pigments/hansa_yellow.png"/> | 252, 211, 0  | 0.988, 0.827, 0.0 | 0.973, 0.651, 0.0 |
| Cadmium Orange | <img src="https://scrtwpns.com/mixbox/pigments/cadmium_orange.png"/> | 255, 105, 0  | 1.0, 0.412, 0.0 | 1.0, 0.141, 0.0 |
| Cadmium Red | <img src="https://scrtwpns.com/mixbox/pigments/cadmium_red.png"/> | 255, 39, 2  | 1.0, 0.153, 0.008 | 1.0, 0.02, 0.001 |
| Quinacridone Magenta | <img src="https://scrtwpns.com/mixbox/pigments/quinacridone_magenta.png"/> | 128, 2, 46  | 0.502, 0.008, 0.18 | 0.216, 0.001, 0.027 |
| Cobalt Violet | <img src="https://scrtwpns.com/mixbox/pigments/cobalt_violet.png"/> | 78, 0, 66  | 0.306, 0.0, 0.259 | 0.076, 0.0, 0.054 |
| Ultramarine Blue | <img src="https://scrtwpns.com/mixbox/pigments/ultramarine_blue.png"/> | 25, 0, 89  | 0.098, 0.0, 0.349 | 0.01, 0.0, 0.1 |
| Cobalt Blue | <img src="https://scrtwpns.com/mixbox/pigments/cobalt_blue.png"/> | 0, 33, 133  | 0.0, 0.129, 0.522 | 0.0, 0.015, 0.235 |
| Phthalo Blue | <img src="https://scrtwpns.com/mixbox/pigments/phthalo_blue.png"/> | 13, 27, 68  | 0.051, 0.106, 0.267 | 0.004, 0.011, 0.058 |
| Phthalo Green | <img src="https://scrtwpns.com/mixbox/pigments/phthalo_green.png"/> | 0, 60, 50  | 0.0, 0.235, 0.196 | 0.0, 0.045, 0.032 |
| Permanent Green | <img src="https://scrtwpns.com/mixbox/pigments/permanent_green.png"/> | 7, 109, 22  | 0.027, 0.427, 0.086 | 0.002, 0.153, 0.008 |
| Sap Green | <img src="https://scrtwpns.com/mixbox/pigments/sap_green.png"/> | 107, 148, 4  | 0.42, 0.58, 0.016 | 0.147, 0.296, 0.001 |
| Burnt Sienna | <img src="https://scrtwpns.com/mixbox/pigments/burnt_sienna.png"/> | 123, 72, 0  | 0.482, 0.282, 0.0 | 0.198, 0.065, 0.0 |