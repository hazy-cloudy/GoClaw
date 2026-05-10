package proactive

import "time"

type InterruptibilityEvaluator struct{}

func NewInterruptibilityEvaluator() *InterruptibilityEvaluator {
	return &InterruptibilityEvaluator{}
}

// Evaluate 只回答一个问题：
// “当前这个用户，适不适合被桌宠主动打扰？”
//
// 这里故意不看事件类型本身，只看时机。
// 这样 weekly_report 和 progress_nudge 都能共用这一层。
func (e *InterruptibilityEvaluator) Evaluate(snapshot Snapshot) InterruptibilityDecision {
	now := snapshot.Now
	if now.IsZero() {
		now = time.Now()
	}

	// 用户当前正在忙，就不要打扰。
	if snapshot.Activity.CurrentSessionBusy {
		if !snapshot.Activity.LastUserMessageAt.IsZero() {
			sinceLastMsg := now.Sub(snapshot.Activity.LastUserMessageAt)
			if sinceLastMsg >= 45*time.Second {
				return InterruptibilityDecision{
					Interruptible: true,
					Level:         InterruptibilitySoft,
					ReasonCodes:   []string{"recent_user_activity"},
					NextCheckAfter: 0,
				}
			}
		}
		return InterruptibilityDecision{
			Interruptible: false,
			Level:         InterruptibilityNo,
			ReasonCodes:   []string{"session_busy"},
			NextCheckAfter: 2 * time.Minute,
		}
	}

	// 刚刚才被主动打扰过，也先别连续轰炸。
	if !snapshot.Activity.LastPushAt.IsZero() && now.Sub(snapshot.Activity.LastPushAt) < 5*time.Minute {
		return InterruptibilityDecision{
			Interruptible: false,
			Level:         InterruptibilityNo,
			ReasonCodes:   []string{"recent_proactive_push"},
			NextCheckAfter: 5*time.Minute - now.Sub(snapshot.Activity.LastPushAt),
		}
	}

	// 用户刚刚连续发很多消息，说明当前更像“任务处理中”，先不要插嘴。
	if !snapshot.Activity.LastUserMessageAt.IsZero() {
		sinceLastMsg := now.Sub(snapshot.Activity.LastUserMessageAt)
		if sinceLastMsg < 30*time.Second && snapshot.Activity.RecentMessageCount >= 3 {
			return InterruptibilityDecision{
				Interruptible: false,
				Level:         InterruptibilityNo,
				ReasonCodes:   []string{"active_user_conversation"},
				NextCheckAfter: 90 * time.Second,
			}
		}
		if sinceLastMsg < 2*time.Minute {
			return InterruptibilityDecision{
				Interruptible: true,
				Level:         InterruptibilitySoft,
				ReasonCodes:   []string{"recent_user_activity"},
				NextCheckAfter: 0,
			}
		}
	}

	// 如果用户当前就看着控制台/桌宠窗口，适合更积极一点地投递。
	if snapshot.Activity.ConsoleVisible || snapshot.Activity.PetVisible {
		return InterruptibilityDecision{
			Interruptible: true,
			Level:         InterruptibilityYes,
			ReasonCodes:   []string{"surface_visible"},
			NextCheckAfter: 0,
		}
	}

	return InterruptibilityDecision{
		Interruptible: true,
		Level:         InterruptibilitySoft,
		ReasonCodes:   []string{"default_idle_state"},
		NextCheckAfter: 0,
	}
}
