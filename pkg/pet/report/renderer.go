package report

import "strings"

// RenderWeeklySummary 负责把结构化周报，渲染成一段更像“桌宠会说的话”。
//
// 这里要刻意区分三种不同语义的字段：
// 1. persona: 角色的长文本设定，适合 prompt，不适合直接做模板分支
// 2. persona_type: 稳定枚举，适合做程序分支（gentle/cool/playful）
// 3. personality_tone: 交互语气标签，适合做“阴阳怪气/抽象发疯/甜心夹子”这种风格判断
//
// 当前这个渲染器的策略是：
// - 优先使用 personalityTone 决定具体语气
// - 如果没有 tone，再退回 personaType 做风格兜底
func RenderWeeklySummary(report WeeklyReport, personaType, personalityTone, dominantEmotion string) string {
	base := strings.TrimSpace(report.Summary)
	if base == "" {
		base = "这周的记录我先帮你收好了。"
	}

	normalizedTone := strings.TrimSpace(personalityTone)
	normalizedType := strings.ToLower(strings.TrimSpace(personaType))

	switch normalizedTone {
	case "阴阳怪气":
		return "这周又把我当全能外包用了是吧。" + base
	case "抽象发疯":
		return "本周精神轨迹已完成重组。" + base
	case "甜心夹子":
		return "这周也陪你跑完一段啦。" + base
	}

	switch normalizedType {
	case "cool":
		return "本周回顾如下。" + base
	case "playful":
		return "我来播报一下这周战况。" + base
	case "gentle":
		return "我把这周的记录整理好了。" + base
	default:
		return "我把这周的记录整理好了。" + base
	}
}
