package render

import (
	"net/url"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// relativeImageRewriter is a goldmark AST transformer that rewrites relative
// image destinations (e.g. "photo.jpg" or "./photo.jpg") to point to the
// gist's raw file endpoint so that images embedded alongside markdown files
// in a gist render correctly.
type relativeImageRewriter struct {
	rawBaseURL string // e.g. "/user/gistid/raw/HEAD"
}

func (r *relativeImageRewriter) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	if r.rawBaseURL == "" {
		return
	}

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		img, ok := n.(*ast.Image)
		if !ok {
			return ast.WalkContinue, nil
		}

		dest := string(img.Destination)
		if dest == "" {
			return ast.WalkContinue, nil
		}

		// Skip absolute URLs and data URIs
		if strings.HasPrefix(dest, "http://") || strings.HasPrefix(dest, "https://") ||
			strings.HasPrefix(dest, "data:") || strings.HasPrefix(dest, "//") {
			return ast.WalkContinue, nil
		}

		// Skip fragment-only references
		if strings.HasPrefix(dest, "#") {
			return ast.WalkContinue, nil
		}

		// Strip leading "./"
		dest = strings.TrimPrefix(dest, "./")

		// Split off any query/fragment
		clean := dest
		suffix := ""
		if idx := strings.IndexAny(clean, "?#"); idx >= 0 {
			suffix = clean[idx:]
			clean = clean[:idx]
		}

		// Build the rewritten URL: rawBaseURL/filename
		rewritten := strings.TrimRight(r.rawBaseURL, "/") + "/" + url.PathEscape(clean) + suffix
		img.Destination = []byte(rewritten)

		return ast.WalkContinue, nil
	})
}
