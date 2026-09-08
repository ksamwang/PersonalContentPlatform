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
