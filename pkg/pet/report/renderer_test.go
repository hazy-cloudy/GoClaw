package report

import "testing"

func TestRenderWeeklySummaryToneFirst(t *testing.T) {
	summary := RenderWeeklySummary(
		WeeklyReport{Summary: "这周你主要在修复和文档上忙。"},
		"gentle",
		"阴阳怪气",
		"joy",
	)
	if summary == "" {
		t.Fatal("summary should not be empty")
	}
	if summary[:len("这周又把我当全能外包用了是吧。")] != "这周又把我当全能外包用了是吧。" {
		t.Fatalf("unexpected summary prefix: %q", summary)
	}
}

func TestRenderWeeklySummaryFallbackToType(t *testing.T) {
	summary := RenderWeeklySummary(
		WeeklyReport{Summary: "这周你主要在修复和文档上忙。"},
		"cool",
		"",
		"joy",
	)
	if summary[:len("本周回顾如下。")] != "本周回顾如下。" {
		t.Fatalf("unexpected summary prefix: %q", summary)
	}
}
