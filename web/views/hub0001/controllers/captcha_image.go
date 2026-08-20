package controllers

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	mathrand "math/rand"
	"time"
)

const (
	captchaImageWidth  = 140
	captchaImageHeight = 50
	captchaGlyphWidth  = 5
	captchaGlyphHeight = 7
	captchaGlyphScale  = 3
)

// 5x7 点阵，仅数字，避免 SVG/字体把答案写进可解析文本。
var digitGlyphs = [10][captchaGlyphHeight]byte{
	{0x0E, 0x11, 0x13, 0x15, 0x19, 0x11, 0x0E}, // 0
	{0x04, 0x0C, 0x04, 0x04, 0x04, 0x04, 0x0E}, // 1
	{0x0E, 0x11, 0x01, 0x02, 0x04, 0x08, 0x1F}, // 2
	{0x0E, 0x11, 0x01, 0x06, 0x01, 0x11, 0x0E}, // 3
	{0x02, 0x06, 0x0A, 0x12, 0x1F, 0x02, 0x02}, // 4
	{0x1F, 0x10, 0x1E, 0x01, 0x01, 0x11, 0x0E}, // 5
	{0x0E, 0x10, 0x1E, 0x11, 0x11, 0x11, 0x0E}, // 6
	{0x1F, 0x01, 0x02, 0x04, 0x08, 0x08, 0x08}, // 7
	{0x0E, 0x11, 0x11, 0x0E, 0x11, 0x11, 0x0E}, // 8
	{0x0E, 0x11, 0x11, 0x0F, 0x01, 0x01, 0x0E}, // 9
}

// renderCaptchaPNGDataURI 将验证码画成 PNG Data URI。点阵绘制，答案不会以文本形式出现在图数据里。
func renderCaptchaPNGDataURI(code string) (string, error) {
	rng := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	img := image.NewNRGBA(image.Rect(0, 0, captchaImageWidth, captchaImageHeight))

	for y := 0; y < captchaImageHeight; y++ {
		for x := 0; x < captchaImageWidth; x++ {
			gray := uint8(230 + rng.Intn(20))
			img.SetNRGBA(x, y, color.NRGBA{R: gray, G: gray, B: gray, A: 255})
		}
	}

	for i := 0; i < 180; i++ {
		img.SetNRGBA(rng.Intn(captchaImageWidth), rng.Intn(captchaImageHeight), randomNoiseColor(rng))
	}

	for i := 0; i < 4; i++ {
		drawLine(img,
			rng.Intn(captchaImageWidth), rng.Intn(captchaImageHeight),
			rng.Intn(captchaImageWidth), rng.Intn(captchaImageHeight),
			color.NRGBA{R: 120, G: 110, B: 160, A: 90})
	}

	digitW := captchaGlyphWidth * captchaGlyphScale
	gap := 6
	totalW := len(code)*digitW + (len(code)-1)*gap
	startX := (captchaImageWidth - totalW) / 2
	if startX < 2 {
		startX = 2
	}

	for i, ch := range code {
		if ch < '0' || ch > '9' {
			continue
		}
		digit := int(ch - '0')
		x := startX + i*(digitW+gap) + rng.Intn(3) - 1
		y := 12 + rng.Intn(7)
		fg := color.NRGBA{
			R: uint8(60 + rng.Intn(50)),
			G: uint8(50 + rng.Intn(40)),
			B: uint8(90 + rng.Intn(50)),
			A: 255,
		}
		drawDigit(img, x, y, digit, fg)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func drawDigit(img *image.NRGBA, originX, originY, digit int, fg color.NRGBA) {
	glyph := digitGlyphs[digit]
	for row := 0; row < captchaGlyphHeight; row++ {
		bits := glyph[row]
		for col := 0; col < captchaGlyphWidth; col++ {
			if bits&(1<<uint(captchaGlyphWidth-1-col)) == 0 {
				continue
			}
			for dy := 0; dy < captchaGlyphScale; dy++ {
				for dx := 0; dx < captchaGlyphScale; dx++ {
					img.SetNRGBA(originX+col*captchaGlyphScale+dx, originY+row*captchaGlyphScale+dy, fg)
				}
			}
		}
	}
}

func drawLine(img *image.NRGBA, x0, y0, x1, y1 int, c color.NRGBA) {
	dx := absInt(x1 - x0)
	dy := -absInt(y1 - y0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		if image.Pt(x0, y0).In(img.Bounds()) {
			img.SetNRGBA(x0, y0, c)
		}
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func randomNoiseColor(rng *mathrand.Rand) color.NRGBA {
	return color.NRGBA{
		R: uint8(150 + rng.Intn(80)),
		G: uint8(150 + rng.Intn(80)),
		B: uint8(160 + rng.Intn(80)),
		A: 180,
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
