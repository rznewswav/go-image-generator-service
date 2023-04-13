package main

import (
	"fmt"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
	"image/color"
	"os"
	"strings"
)

const mmPerPixel = 0.2645833333

const defaultFontPath = "resources/fonts/nunito-v23-latin-regular.ttf"

var loadedFonts = make(map[string]*canvas.Font)

func LoadFont(path ...string) *canvas.Font {
	var fontPath = defaultFontPath
	if len(path) > 0 {
		fontPath = path[0]
	}
	if len(path) > 1 {
		fmt.Printf("cannot load multiple fonts at once, loading only %s, ignoring: %s", fontPath, strings.Join(path[1:], ", "))
	}

	if alreadyLoaded, hasLoadedBefore := loadedFonts[fontPath]; hasLoadedBefore {
		return alreadyLoaded
	}

	font, fontError := canvas.LoadFontFile(fontPath, canvas.FontRegular)
	if fontError != nil {
		panic(fontError)
	}

	loadedFonts[fontPath] = font

	return font
}

func GetFirstNonOverflowingText(
	text string,
	width float64,
	maxHeight float64,
	opts ...interface{},
) (canvasText *canvas.Text) {
	var fillColor = canvas.Black
	var halign = canvas.Top
	var maxFontSize = 70

	for _, opt := range opts {
		switch casted := opt.(type) {
		case color.RGBA:
			fillColor = casted
		case canvas.TextAlign:
			halign = casted
		case int:
			maxFontSize = casted
		}

	}

	if maxFontSize < 10 {
		panic(fmt.Errorf("cannot write in font size less than 10pt: %d", maxFontSize))
	}

	for sizePt := maxFontSize; sizePt >= 10; sizePt-- {
		face := LoadFont().Face(float64(sizePt), fillColor)
		canvasText := canvas.NewTextBox(face, text, width, 0, halign, canvas.Left, .0, .0)
		if canvasText.OutlineBounds().H < maxHeight {
			return canvasText
		}
	}
	return canvasText
}

type TextConfig struct {
	VariableName string
	RelativeX    float64
	RelativeY    float64
	FontSize     int
	MaxWidth     float64
	MaxHeight    float64
	TextAlign    canvas.TextAlign
	Color        color.RGBA
}

type ImageTemplate struct {
	ResourcePath string
	Config       []TextConfig
}

var weeklyTemplate = ImageTemplate{
	"resources/images/Confirmed (Weekly & Total).png",
	[]TextConfig{
		{
			"{weekly new cases}",
			0.5,
			0.44,
			55,
			128,
			18,
			canvas.Center,
			canvas.Black,
		},
		{
			"{weekly total cases}",
			0.5,
			0.20,
			55,
			128,
			18,
			canvas.Center,
			canvas.Black,
		},
	},
}

type M = map[string]string

func Run(
	template ImageTemplate,
	outPath string,
	variableValues M,
) {
	fileHandler, fileReadError := os.Open(template.ResourcePath)
	if fileReadError != nil {
		panic(fileReadError)
	}

	textConfigs := template.Config

	pngImage, pngImageError := canvas.NewPNGImage(fileHandler)
	if pngImageError != nil {
		panic(pngImageError)
	}

	pngSize := pngImage.Image.Bounds().Max

	c := canvas.NewFromSize(canvas.Size{
		W: float64(pngSize.X) * mmPerPixel,
		H: float64(pngSize.Y) * mmPerPixel,
	})

	ctx := canvas.NewContext(c)

	ctx.DrawImage(0, 0, pngImage, canvas.DefaultResolution)

	for _, config := range textConfigs {
		var text = config.VariableName
		if value, variableInValueMap := variableValues[config.VariableName]; variableInValueMap {
			text = value
		}
		textBox := GetFirstNonOverflowingText(text, config.MaxWidth, config.MaxHeight, config.FontSize, config.Color, config.TextAlign)
		var x float64
		var y float64
		var maxWidth = config.MaxWidth
		if textBox.Bounds().W > maxWidth {
			maxWidth = textBox.Bounds().W
		}
		switch config.TextAlign {
		case canvas.Center:
			{
				x = c.W*config.RelativeX - maxWidth*0.5
				y = c.H*config.RelativeY + textBox.Bounds().H*0.5
			}
		case canvas.Right:
			{
				x = c.W*config.RelativeX - maxWidth
				y = c.H*config.RelativeY + textBox.Bounds().H*0.5
			}
		default:
			{
				x = c.W * config.RelativeX
				y = c.H*config.RelativeY + textBox.Bounds().H*0.5
			}

		}

		ctx.SetStrokeWidth(2)
		ctx.SetFillColor(canvas.Lightblue)
		ctx.MoveTo(x, y)
		ctx.LineTo(x+textBox.Bounds().W, y)
		ctx.LineTo(x+textBox.Bounds().W, y-textBox.Bounds().H)
		ctx.LineTo(x, y-textBox.Bounds().H)
		ctx.LineTo(x, y)
		ctx.Close()
		ctx.Stroke()

		ctx.DrawText(
			x,
			y,
			textBox,
		)
	}

	writer := renderers.PNG(canvas.DPMM(3.2))
	writerError := c.WriteFile(outPath, writer)
	if writerError != nil {
		panic(writerError)
	}
}

func main() {
	Run(weeklyTemplate, "/tmp/out.png", M{
		"{weekly new cases}": "501,784",
	})
}
