package document

import (
	"strings"
	"testing"
)

func TestHTMLRendersSupportedNodesAndEscapesText(t *testing.T) {
	input := []byte(`{"schemaVersion":1,"type":"doc","content":[{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"Heading"}]},{"type":"paragraph","content":[{"type":"text","text":"<script>alert(1)</script>"}]}]}`)
	output := HTML(input)
	if !strings.Contains(output, "<h2>Heading</h2>") {
		t.Fatalf("missing heading: %s", output)
	}
	if strings.Contains(output, "<script>") {
		t.Fatalf("unsafe markup was not escaped: %s", output)
	}
}
func TestPlainTextFlattensDocument(t *testing.T) {
	input := []byte(`{"schemaVersion":1,"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"hello"},{"type":"text","text":"world"}]}]}`)
	if got := PlainText(input); got != "hello world" {
		t.Fatalf("got %q", got)
	}
}

func TestHTMLRendersRichContent(t *testing.T) {
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Read","marks":[{"type":"bold"},{"type":"link","attrs":{"href":"https://example.com"}}]}]},{"type":"image","attrs":{"src":"/media/123","alt":"示例图片"}},{"type":"table","content":[{"type":"tableRow","content":[{"type":"tableHeader","content":[{"type":"paragraph","content":[{"type":"text","text":"标题"}]}]}]}]}]}`)
	got := HTML(input)
	for _, want := range []string{`<a href="https://example.com"><strong>Read</strong></a>`, `<img src="/media/123" alt="示例图片">`, `<table>`} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered HTML %q does not contain %q", got, want)
		}
	}
}
