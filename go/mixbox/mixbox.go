
package mixbox

import (
	"math"
	"os"
	"bytes"
	"compress/flate"
	"encoding/base64"
	"fmt"
	"io"
)

const LatentSize = 7

var lut []uint8

func InitLUT(lutData []uint8) {
	lut = lutData
}

func LoadLUTFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read LUT file: %w", err)
	}
	InitLUT(data)
	return nil
}


func clamp01(x float64) float64 {
	if x < 0.0 {
		return 0.0
	}
	if x > 1.0 {
		return 1.0
	}
	return x
}

func srgbToLinear(x float64) float64 {
	if x >= 0.04045 {
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return x / 12.92
}

func linearToSrgb(x float64) float64 {
	if x >= 0.0031308 {
		return 1.055*math.Pow(x, 1.0/2.4) - 0.055
	}
	return 12.92 * x
}

func evalPolynomial(c0, c1, c2, c3 float64) [3]float64 {
	c00 := c0 * c0
	c11 := c1 * c1
	c22 := c2 * c2
	c33 := c3 * c3
	c01 := c0 * c1
	c02 := c0 * c2
	c12 := c1 * c2

	var r, g, b float64

	r += 0.07717053 * c0 * c00
	g += 0.02826978 * c0 * c00
	b += 0.24832992 * c0 * c00

	r += 0.95912302 * c1 * c11
	g += 0.80256528 * c1 * c11
	b += 0.03561839 * c1 * c11

	r += 0.74683774 * c2 * c22
	g += 0.04868586 * c2 * c22

	r += 0.99518138 * c3 * c33
	g += 0.99978149 * c3 * c33
	b += 0.99704802 * c3 * c33

	r += 0.04819146 * c00 * c1
	g += 0.83363781 * c00 * c1
	b += 0.32515377 * c00 * c1

	r += -0.68146950 * c01 * c1
	g += 1.46107803 * c01 * c1
	b += 1.06980936 * c01 * c1

	r += 0.27058419 * c00 * c2
	g += -0.15324870 * c00 * c2
	b += 1.98735057 * c00 * c2

	r += 0.80478189 * c02 * c2
	g += 0.67093710 * c02 * c2
	b += 0.18424500 * c02 * c2

	r += -0.35031003 * c00 * c3
	g += 1.37855826 * c00 * c3
	b += 3.68865000 * c00 * c3

	r += 1.05128046 * c0 * c33
	g += 1.97815239 * c0 * c33
	b += 2.82989073 * c0 * c33

	r += 3.21607125 * c11 * c2
	g += 0.81270228 * c11 * c2
	b += 1.03384539 * c11 * c2

	r += 2.78893374 * c1 * c22
	g += 0.41565549 * c1 * c22
	b += -0.04487295 * c1 * c22

	r += 3.02162577 * c11 * c3
	g += 2.55374103 * c11 * c3
	b += 0.32766114 * c11 * c3

	r += 2.95124691 * c1 * c33
	g += 2.81201112 * c1 * c33
	b += 1.17578442 * c1 * c33

	r += 2.82677043 * c22 * c3
	g += 0.79933038 * c22 * c3
	b += 1.81715262 * c22 * c3

	r += 2.99691099 * c2 * c33
	g += 1.22593053 * c2 * c33
	b += 1.80653661 * c2 * c33

	r += 1.87394106 * c01 * c2
	g += 2.05027182 * c01 * c2
	b += -0.29835996 * c01 * c2

	r += 2.56609566 * c01 * c3
	g += 7.03428198 * c01 * c3
	b += 0.62575374 * c01 * c3

	r += 4.08329484 * c02 * c3
	g += -1.40408358 * c02 * c3
	b += 2.14995522 * c02 * c3

	r += 6.00078678 * c12 * c3
	g += 2.55552042 * c12 * c3
	b += 1.90739502 * c12 * c3

	return [3]float64{r, g, b}
}

func RGBToLatent(rgb [3]uint8) [LatentSize]float64 {
	return FloatRGBToLatent([3]float64{
		float64(rgb[0]) / 255.0,
		float64(rgb[1]) / 255.0,
		float64(rgb[2]) / 255.0,
	})
}

func FloatRGBToLatent(rgb [3]float64) [LatentSize]float64 {
	r := clamp01(rgb[0])
	g := clamp01(rgb[1])
	b := clamp01(rgb[2])

	x := r * 63.0
	y := g * 63.0
	z := b * 63.0

	ix := int(x)
	iy := int(y)
	iz := int(z)

	tx := x - float64(ix)
	ty := y - float64(iy)
	tz := z - float64(iz)

	xyz := (ix + iy*64 + iz*64*64) & 0x3FFFF

	var c0, c1, c2 float64

	get := func(offset int) float64 {
		return float64(lut[xyz+offset])
	}

	weights := []float64{
		(1.0 - tx) * (1.0 - ty) * (1.0 - tz),
		tx * (1.0 - ty) * (1.0 - tz),
		(1.0 - tx) * ty * (1.0 - tz),
		tx * ty * (1.0 - tz),
		(1.0 - tx) * (1.0 - ty) * tz,
		tx * (1.0 - ty) * tz,
		(1.0 - tx) * ty * tz,
		tx * ty * tz,
	}

	offsets := []int{192, 193, 256, 257, 4288, 4289, 4352, 4353}

	for i := 0; i < 8; i++ {
		w := weights[i]
		c0 += w * get(offsets[i])
		c1 += w * get(offsets[i] + 262144)
		c2 += w * get(offsets[i] + 524288)
	}

	c0 /= 255.0
	c1 /= 255.0
	c2 /= 255.0

	c3 := 1.0 - (c0 + c1 + c2)

	mix := evalPolynomial(c0, c1, c2, c3)

	return [LatentSize]float64{
		c0, c1, c2, c3,
		r - mix[0],
		g - mix[1],
		b - mix[2],
	}
}

func LatentToRGB(latent [LatentSize]float64) [3]uint8 {
	rgb := evalPolynomial(latent[0], latent[1], latent[2], latent[3])
	return [3]uint8{
		uint8(clamp01(rgb[0]+latent[4])*255.0 + 0.5),
		uint8(clamp01(rgb[1]+latent[5])*255.0 + 0.5),
		uint8(clamp01(rgb[2]+latent[6])*255.0 + 0.5),
	}
}

func Lerp(rgb1, rgb2 [3]uint8, t float64) [3]uint8 {
	latent1 := RGBToLatent(rgb1)
	latent2 := RGBToLatent(rgb2)

	var mixed [LatentSize]float64
	for i := 0; i < LatentSize; i++ {
		mixed[i] = (1.0 - t) * latent1[i] + t * latent2[i]
	}

	return LatentToRGB(mixed)
}

// decompress takes a raw base64-encoded, deflate-compressed string and returns the decompressed LUT data.
func decompress(input string) ([]byte, error) {
	// Step 1: Base64 decode the input string.
	data, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %v", err)
	}

	// Step 2: Decompress the raw deflate data using compress/flate.
	reader := flate.NewReader(bytes.NewReader(data))
	defer reader.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return nil, fmt.Errorf("flate decompression failed: %v", err)
	}
	output := buf.Bytes()

	// Step 3: Process the output bytes as in the Python version:
	// For each index i, output[i] = (output[i-1] if (i & 63) != 0 else 127) + (output[i] - 127)
	for i := 0; i < len(output); i++ {
		var prev byte
		if i&63 != 0 {
			prev = output[i-1]
		} else {
			prev = 127
		}
		output[i] = prev + (output[i] - 127)
	}

	// Step 4: Append 4161 zeros.
	output = append(output, make([]byte, 4161)...)

	return output, nil
}

// DecompressAndInitLUT decompresses the LUT from a raw string and initializes the global LUT.
func DecompressAndInitLUT(lutString string) error {
	data, err := decompress(lutString)
	if err != nil {
		return err
	}
	InitLUT(data)
	return nil
}