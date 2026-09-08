package document

import (
	"bytes"
	"encoding/json"
	"html"
	"strings"
)

type Node struct {
	Type    string         `json:"type"`
	Text    string         `json:"text,omitempty"`
	Attrs   map[string]any `json:"attrs,omitempty"`
	Content []Node         `json:"content,omitempty"`
}
type Document struct {
	SchemaVersion int    `json:"schemaVersion"`
	Type          string `json:"type"`
	Content       []Node `json:"content"`
}

func Parse(data []byte) (Document, error) {
	var doc Document
	err := json.Unmarshal(data, &doc)
	return doc, err
}
func PlainText(data []byte) string {
	doc, err := Parse(data)
	if err != nil {
		return ""
	}
	var out []string
	walk(doc.Content, func(n Node) {
		if n.Text != "" {
			out = append(out, n.Text)
		}
	})
	return strings.Join(out, " ")
}
func HTML(data []byte) string {
	doc, err := Parse(data)
	if err != nil {
		return ""
	}
	var out bytes.Buffer
	for _, node := range doc.Content {
		renderNode(&out, node)
	}
	return out.String()
}
func walk(nodes []Node, fn func(Node)) {
	for _, n := range nodes {
		fn(n)
		walk(n.Content, fn)
	}
}
func renderNode(out *bytes.Buffer, n Node) {
	text := func() {
		for _, child := range n.Content {
			if child.Type == "text" {
				out.WriteString(html.EscapeString(child.Text))
			} else {
				renderNode(out, child)
			}
		}
	}
	switch n.Type {
	case "paragraph":
		out.WriteString("<p>")
		text()
		out.WriteString("</p>")
	case "heading":
		level := 1
		if v, ok := n.Attrs["level"].(float64); ok && v >= 1 && v <= 6 {
			level = int(v)
		}
		out.WriteString("<h" + string(rune('0'+level)) + ">")
		text()
		out.WriteString("</h" + string(rune('0'+level)) + ">")
	case "blockquote":
		out.WriteString("<blockquote>")
		text()
		out.WriteString("</blockquote>")
	case "bulletList":
		out.WriteString("<ul>")
		for _, c := range n.Content {
			renderNode(out, c)
		}
		out.WriteString("</ul>")
	case "orderedList":
		out.WriteString("<ol>")
		for _, c := range n.Content {
			renderNode(out, c)
		}
		out.WriteString("</ol>")
	case "listItem":
		out.WriteString("<li>")
		text()
		out.WriteString("</li>")
	case "codeBlock":
		out.WriteString("<pre><code>")
		text()
		out.WriteString("</code></pre>")
	case "horizontalRule":
		out.WriteString("<hr>")
	case "callout":
		out.WriteString(`<aside class="callout">`)
		text()
		out.WriteString("</aside>")
	case "text":
		out.WriteString(html.EscapeString(n.Text))
	default:
		out.WriteString(`<div data-unsupported-block="` + html.EscapeString(n.Type) + `"></div>`)
	}
}
