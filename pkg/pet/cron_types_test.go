package pet

import "testing"

func TestCronActionConstants(t *testing.T) {
	if ActionCronAdd != "cron_add" {
		t.Fatalf("ActionCronAdd = %q", ActionCronAdd)
	}
	if ActionCronList != "cron_list" {
		t.Fatalf("ActionCronList = %q", ActionCronList)
	}
	if ActionCronUpdate != "cron_update" {
		t.Fatalf("ActionCronUpdate = %q, want cron_update", ActionCronUpdate)
	}
	if ActionCronRemove != "cron_remove" {
		t.Fatalf("ActionCronRemove = %q", ActionCronRemove)
	}
}

func TestCronRequestAndResponseTypes(t *testing.T) {
	req := CronAddRequest{
		Name:         "喝水提醒",
		Description:  "10分钟后提醒我喝水",
		ScheduleType: "at",
		Message:      "10分钟后提醒我喝水",
		Command:      "echo hi",
	}
	if req.Description == "" {
		t.Fatal("expected description to be available")
	}
	if req.ScheduleType != "at" {
		t.Fatalf("ScheduleType = %q, want at", req.ScheduleType)
	}

	resp := CronJobInfo{
		ID:           "job-1",
		Name:         "喝水提醒",
		Description:  "10分钟后提醒我喝水",
		Message:      "10分钟后提醒我喝水",
		ScheduleType: "at",
		ScheduleKind: "at",
		Schedule:     "at:123",
		Command:      "echo hi",
		LastError:    "",
		UpdatedAtMS:  123,
	}
	if resp.Description == "" {
		t.Fatal("expected description to be available")
	}
	if resp.ScheduleType != resp.ScheduleKind {
		t.Fatalf("ScheduleType = %q, ScheduleKind = %q, want aligned", resp.ScheduleType, resp.ScheduleKind)
	}
}
