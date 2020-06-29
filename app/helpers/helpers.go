package helpers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"image/color"
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"gallo/app/models"
	"gallo/lib"
)

type AssetsHashLookup map[string]string

var assetsHashLookup AssetsHashLookup

var Funcs template.FuncMap

func SrcSetSizes(image models.Image) (string, error) {
	tmplString := `srcset="{{ .SrcSet }}" sizes="{{ .Sizes }}"`
	tmpl, err := template.New("img").Parse(tmplString)
	if err != nil {
		return "", err
	}
	previews := image.GetPreviews()
	srcSet := make([]string, len(previews))

	for i := range previews {
		srcSet[i] = fmt.Sprintf("%s %dw", previews[i].URL, previews[i].Width)
	}

	data := struct {
		SrcSet string
		Sizes  string
	}{
		strings.Join(srcSet, ", "),
		"(max-width: 320px) 100vw, (max-width: 630px) 50vw, 33vw",
	}

	var str strings.Builder
	tmpl.Execute(&str, data)

	return str.String(), nil
}

// Shrinks the width and height of an Image as much as possible without changing
// the aspect ratio
//
// The reason for doing this, is because large values for width and height in
// the generated svg, seem to prohibit proper rendering on Safari iOS 5.1.
// Reducing width/height ratio to lowest common terms, appears to solve the
// issue.
func shrink(image models.Image) (width, height int) {
	// fisher/yates
	var gcd = func(n, m int) int {
		for m != 0 {
			n, m = m, n%m
		}

		return n
	}

	divisor := gcd(image.GetWidth(), image.GetHeight())

	width = image.GetWidth() / divisor
	height = image.GetHeight() / divisor

	return
}

func NewAssetsHashLookup(digestPaths ...string) AssetsHashLookup {
	lookup := make(AssetsHashLookup)

	for _, digestPath := range digestPaths {
		file, err := os.Open(digestPath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		var (
			hash     string
			fileName string
		)

		for {
			_, err := fmt.Fscanf(file, "%s %s", &hash, &fileName)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			lookup[fileName] = hash
		}

	}
	return lookup
}

func init() {
	appPath := lib.MustGetEnv("APP_PATH")
	appEnv := lib.MustGetEnv("APP_ENV")
	appVersion := lib.MustGetEnv("APP_VERSION")

	assetsHashLookup = NewAssetsHashLookup(
		path.Join(appPath, "/public/assets/css/sha256sum.txt"),
		path.Join(appPath, "/public/assets/js/sha256sum.txt"),
	)

	Funcs = template.FuncMap{
		"safeHTMLAttr": func(s string) template.HTMLAttr {
			return template.HTMLAttr(s)
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},
		"pathTo": func(model models.Model) string {
			if model == nil {
				log.Fatal("pathTo: model is Nil")
			}

			return path.Join("/", model.PluralName(), model.ID())
		},
		"pathToCss": func(fileName string) string {
			value := path.Join("/assets/css", fileName)

			if appEnv == "production" {
				hash, ok := assetsHashLookup[fileName]

				if !ok {
					log.Fatalf("Unknown css file: %s", fileName)
				}

				fileExt := filepath.Ext(fileName)
				fileWithoutExt := strings.TrimSuffix(fileName, fileExt)
				newFileName := fmt.Sprintf("%s.%s%s", fileWithoutExt, hash, fileExt)

				value = path.Join("/assets/css", newFileName)
			}

			return value
		},
		"pathToJs": func(fileName string) string {
			value := path.Join("/assets/js", fileName)

			if appEnv == "production" {
				hash, ok := assetsHashLookup[fileName]

				if !ok {
					log.Fatalf("Unknown js file: %s", fileName)
				}

				fileExt := filepath.Ext(fileName)
				fileWithoutExt := strings.TrimSuffix(fileName, fileExt)
				newFileName := fmt.Sprintf("%s.%s%s", fileWithoutExt, hash, fileExt)

				value = path.Join("/assets/js", newFileName)
			}

			return value
		},
		"srcSetSizes": SrcSetSizes,
		"placeholderURI": func(image models.Image) string {
			width, height := shrink(image)

			placeHolder := Placeholder{width, height, image.EdgeColor}

			uri, err := placeHolder.DataURI()
			if err != nil {
				log.Println(err)
				return ""
			}

			return uri
		},
		"toJSON": func(data interface{}) string {
			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Println(err)
				return "{}"
			}

			return string(jsonData)
		},
		"hasField": func(v interface{}, name string) bool {
			rv := reflect.ValueOf(v)
			if rv.Kind() == reflect.Ptr {
				rv = rv.Elem()
			}
			if rv.Kind() != reflect.Struct {
				return false
			}
			return rv.FieldByName(name).IsValid()
		},
		"formatTime": func(t *time.Time) (formatted string) {
			if t == nil {
				return "&nbsp;"
			}

			return t.Format("02 January 2006")
		},
		"repeat": func(n int) []struct{} {
			return make([]struct{}, n)
		},
		"random": func(n int) []struct{} {
			return make([]struct{}, 3+rand.Intn(n))
		},
		"boardBackground": func(board models.Board) (attr string) {
			if board.TrelloBoard.Prefs.BackgroundColor != "" {
				attr = fmt.Sprintf("style=\"background: %s;\"", board.TrelloBoard.Prefs.BackgroundColor)
			} else {
				abs := func(n int) int {
					if n < 0 {
						return -n
					}

					return n
				}

				tmp, desiredWidth, j := 344, 344, 0
				for i := range board.TrelloBoard.Prefs.BackgroundImageScaled {
					diff := abs(desiredWidth - board.TrelloBoard.Prefs.BackgroundImageScaled[i].Width)

					if diff < tmp {
						tmp = diff
						j = i
					}
				}

				backgroundImageURL := board.TrelloBoard.Prefs.BackgroundImageScaled[j].URL
				attr = fmt.Sprintf("style=\"background-image: url(%s);\"", backgroundImageURL)
			}

			return
		},
		"shuffleIcon": func(brightness string) template.HTML {
			tag := fmt.Sprintf(`<img src="/assets/icons/shuffle.%s.svg" alt="Shuffle" class="icon">`, brightness)

			return template.HTML(tag)
		},
		// Calculates whether a color is "light" or "dark" and returns the result
		"colorType": func(stringColor string) (string, error) {
			var c color.RGBA

			_, err := fmt.Sscanf(stringColor, "#%02x%02x%02x", &c.R, &c.G, &c.B)
			if err != nil {
				return "", err
			}

			// Counting the perceptive luminance - human eye favors green color...
			luminance := (0.299*float32(c.R) + 0.587*float32(c.G) + 0.114*float32(c.B)) / 255

			if luminance > 0.5 {
				return "light", nil
			} else {
				return "dark", nil
			}
		},
		"appVersion": func() string {
			return appVersion
		},
	}
}
