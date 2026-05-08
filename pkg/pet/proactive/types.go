package proactive

import "time"

// InterruptibilityLevel 用来表达“用户当前被打扰的适宜程度”。
// 这里故意只分 3 档，避免第一阶段过度设计。
type InterruptibilityLevel string

const (
	InterruptibilityNo   InterruptibilityLevel = "no"
	InterruptibilitySoft InterruptibilityLevel = "soft"
	InterruptibilityYes  InterruptibilityLevel = "yes"
)

type DeliveryLevel string

const (
	DeliveryBlocked DeliveryLevel = "blocked"
	DeliveryBubble  DeliveryLevel = "bubble"
	DeliveryCard    DeliveryLevel = "card"
)

// Snapshot 是主动性系统每次评估时看到的“当前世界状态”。
//
// 它的作用是把原本分散在 pet / userprofile / activity / config 里的信息
// 收敛成一个统一输入，避免每个事件 provider 自己去东拼西凑状态。
type Snapshot struct {
	Now time.Time `json:"now"`

	Pet struct {
		CharacterID     string `json:"character_id"`
		PersonaType     string `json:"persona_type"`
		PersonalityTone string `json:"personality_tone"`
		DominantEmotion string `json:"dominant_emotion"`
		EmotionScore    int    `json:"emotion_score"`
	} `json:"pet"`

	User struct {
		DisplayName     string `json:"display_name"`
		Chronotype      string `json:"chronotype"`
		PersonalityTone string `json:"personality_tone"`
		PressureLevel   string `json:"pressure_level"`
		CurrentMood     string `json:"current_mood"`
		EnergyLevel     int    `json:"energy_level"`
		EngagementLevel int    `json:"engagement_level"`
		StressTrend     string `json:"stress_trend"`
	} `json:"user"`

	Activity struct {
		LastUserMessageAt   time.Time `json:"last_user_message_at"`
		LastPushAt          time.Time `json:"last_push_at"`
		RecentMessageCount  int       `json:"recent_message_count"`
		RecentTaskCount     int       `json:"recent_task_count"`
		UnfinishedTaskCount int       `json:"unfinished_task_count"`
		CurrentSessionBusy  bool      `json:"current_session_busy"`
		ConsoleVisible      bool      `json:"console_visible"`
		PetVisible          bool      `json:"pet_visible"`
	} `json:"activity"`

	Preferences struct {
		ProactiveCare         bool `json:"proactive_care"`
		ProactiveIntervalMins int  `json:"proactive_interval_minutes"`
		WeeklyReportEnabled   bool `json:"weekly_report_enabled"`
		ProgressNudgeEnabled  bool `json:"progress_nudge_enabled"`
		ProactiveCheckMinutes int  `json:"proactive_check_minutes"`
		GlobalCooldownMins    int  `json:"global_cooldown_minutes"`
	} `json:"preferences"`
}

// InterruptibilityDecision 专门回答一个问题：
// “这个用户现在适不适合被打扰？”
type InterruptibilityDecision struct {
	Interruptible bool                  `json:"interruptible"`
	Level         InterruptibilityLevel `json:"level"`
	ReasonCodes   []string              `json:"reason_codes"`
	NextCheckAfter time.Duration        `json:"next_check_after"`
}

// PolicyDecision 则回答另一个问题：
// “在系统规则上，这次主动事件允不允许继续往下走？”
//
// 这和 Interruptibility 是两层不同判断：
// - Interruptibility：用户此刻合不合适被打扰
// - Policy：这次事件是否符合全局冷却、优先级等约束
type PolicyDecision struct {
	Allowed       bool          `json:"allowed"`
	DeliveryLevel DeliveryLevel `json:"delivery_level"`
	Score         int           `json:"score"`
	ReasonCodes   []string      `json:"reason_codes"`
}

// DeliveryHistoryRecord 记录一次“已经成功投递出去”的主动事件，
// 后面做冷却、防重时会用到它。
type DeliveryHistoryRecord struct {
	EventType   string    `json:"event_type"`
	EventID     string    `json:"event_id"`
	CharacterID string    `json:"character_id"`
	DeliveredAt time.Time `json:"delivered_at"`
}
