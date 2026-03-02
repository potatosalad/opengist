package render

import (
	"bytes"
	"regexp"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/thomiceli/opengist/internal/db"
	"github.com/thomiceli/opengist/internal/git"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
	"go.abhg.dev/goldmark/mermaid"
)

func MarkdownGistPreview(gist *db.Gist) (RenderedGist, error) {
	var buf bytes.Buffer
	err := newMarkdown("").Convert([]byte(gist.Preview), &buf)

	// remove links in Markdown Preview, quick fix for now
	re := regexp.MustCompile(`<a\b[^>]*>(.*?)</a>`)
	return RenderedGist{
		Gist: gist,
		HTML: re.ReplaceAllString(buf.String(), `$1`),
	}, err
}

func renderMarkdownFile(file *git.File, rawBaseURL string) (HighlightedFile, error) {
	var buf bytes.Buffer
	err := newMarkdown(rawBaseURL, &svgToImgBase64{}).Convert([]byte(file.Content), &buf)

	return HighlightedFile{
		File: file,
		HTML: buf.String(),
		Type: "Markdown",
	}, err
}

func MarkdownString(content string) (string, error) {
	var buf bytes.Buffer
	err := newMarkdown("", &svgToImgBase64{}).Convert([]byte(content), &buf)

	return buf.String(), err
}

func newMarkdown(rawBaseURL string, extraExtensions ...goldmark.Extender) goldmark.Markdown {
	extensions := []goldmark.Extender{
		extension.GFM,
		highlighting.NewHighlighting(
			highlighting.WithStyle("catppuccin-latte"),
			highlighting.WithFormatOptions(html.WithClasses(true)),
		),
		emoji.Emoji,
		&mermaid.Extender{},
	}

	extensions = append(extensions, extraExtensions...)

	transformers := []util.PrioritizedValue{
		util.Prioritized(&checkboxTransformer{}, 10000),
	}
	if rawBaseURL != "" {
		transformers = append(transformers, util.Prioritized(&relativeImageRewriter{rawBaseURL: rawBaseURL}, 9999))
	}

	return goldmark.New(
		goldmark.WithExtensions(extensions...),
		goldmark.WithParserOptions(
			parser.WithASTTransformers(transformers...),
		),
	)
}
