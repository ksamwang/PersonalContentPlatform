package document

import (
	"bytes"
	"encoding/json"
	"html"
	"net/url"
	"strings"
)

type Mark struct {
	Type  string         `json:"type"`
	Attrs map[string]any `json:"attrs,omitempty"`
}
type Node struct {
	Type    string         `json:"type"`
	Text    string         `json:"text,omitempty"`
	Attrs   map[string]any `json:"attrs,omitempty"`
	Content []Node         `json:"content,omitempty"`
	Marks   []Mark         `json:"marks,omitempty"`
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
func children(out *bytes.Buffer, n Node) {
	for _, child := range n.Content {
		renderNode(out, child)
	}
}
func renderNode(out *bytes.Buffer, n Node) {
	switch n.Type {
	case "text":
		value := html.EscapeString(n.Text)
		for _, mark := range n.Marks {
			switch mark.Type {
			case "bold":
				value = "<strong>" + value + "</strong>"
			case "italic":
				value = "<em>" + value + "</em>"
			case "strike":
				value = "<s>" + value + "</s>"
			case "underline":
				value = "<u>" + value + "</u>"
			case "code":
				value = "<code>" + value + "</code>"
			case "link":
				if href := safeURL(mark.Attrs["href"]); href != "" {
					value = `<a href="` + html.EscapeString(href) + `">` + value + "</a>"
				}
			}
		}
		out.WriteString(value)
	case "paragraph":
		out.WriteString("<p>")
		children(out, n)
		out.WriteString("</p>")
	case "heading":
		level := 1
		if value, ok := n.Attrs["level"].(float64); ok && value >= 1 && value <= 6 {
			level = int(value)
		}
		tag := string(rune('0' + level))
		out.WriteString("<h" + tag + ">")
		children(out, n)
		out.WriteString("</h" + tag + ">")
	case "blockquote":
		out.WriteString("<blockquote>")
		children(out, n)
		out.WriteString("</blockquote>")
	case "bulletList":
		out.WriteString("<ul>")
		children(out, n)
		out.WriteString("</ul>")
	case "orderedList":
		out.WriteString("<ol>")
		children(out, n)
		out.WriteString("</ol>")
	case "taskList":
		out.WriteString(`<ul data-type="taskList">`)
		children(out, n)
		out.WriteString("</ul>")
	case "listItem", "taskItem":
		out.WriteString("<li>")
		children(out, n)
		out.WriteString("</li>")
	case "codeBlock":
		out.WriteString("<pre><code>")
		for _, child := range n.Content {
			out.WriteString(html.EscapeString(child.Text))
		}
		out.WriteString("</code></pre>")
	case "horizontalRule":
		out.WriteString("<hr>")
	case "hardBreak":
		out.WriteString("<br>")
	case "image":
		if src := safeURL(n.Attrs["src"]); src != "" {
			out.WriteString(`<img src="` + html.EscapeString(src) + `" alt="` + html.EscapeString(stringAttr(n.Attrs["alt"])) + `">`)
		}
	case "table":
		out.WriteString("<table>")
		children(out, n)
		out.WriteString("</table>")
	case "tableRow":
		out.WriteString("<tr>")
		children(out, n)
		out.WriteString("</tr>")
	case "tableHeader":
		out.WriteString("<th>")
		children(out, n)
		out.WriteString("</th>")
	case "tableCell":
		out.WriteString("<td>")
		children(out, n)
		out.WriteString("</td>")
	}
}
func stringAttr(value any) string { result, _ := value.(string); return result }
func safeURL(value any) string {
	raw := strings.TrimSpace(stringAttr(value))
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "/media/") {
		return raw
	}
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ""
	}
	return raw
}
