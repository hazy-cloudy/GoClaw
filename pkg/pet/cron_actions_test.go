package pet

import (
	"encoding/json"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/sipeed/picoclaw/pkg/cron"
)

type responseCollector struct {
	mu        sync.Mutex
	responses []Response
}

func (c *responseCollector) push(item any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if resp, ok := item.(Response); ok {
		c.responses = append(c.responses, resp)
	}
}

func (c *responseCollector) popLast(t *testing.T) Response {
	t.Helper()
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.responses) == 0 {
		t.Fatal("expected at least one response")
	}
	resp := c.responses[len(c.responses)-1]
	c.responses = nil
	return resp
}

func setupPetCronService(t *testing.T) (*PetService, *responseCollector) {
	t.Helper()

	workspace := t.TempDir()
	cronStorePath := filepath.Join(workspace, "cron", "jobs.json")
	cronService := cron.NewCronService(cronStorePath, nil)

	service := &PetService{
		config: PetServiceConfig{
			WorkspacePath: workspace,
		},
		cronService:     cronService,
		connSessions:    map[string]string{"conn-1": "session-1"},
		activeSessionID: "session-1",
	}
	service.config.WorkspacePath = workspace
	collector := &responseCollector{}
	service.SetPushHandler(collector.push)
	return service, collector
}

func rawJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	return data
}

func TestPetCronActionsRoundTrip(t *testing.T) {
	service, collector := setupPetCronService(t)

	atMS := time.Now().Add(10 * time.Minute).UnixMilli()
	addReq := Request{
		Action: ActionCronAdd,
		Data: rawJSON(t, CronAddRequest{
			Name:         "喝水提醒",
			Description:  "10分钟后提醒我喝水",
			ScheduleType: "at",
			AtMS:         &atMS,
			Command:      "echo hydrate",
			Enabled:      boolPtr(true),
		}),
	}
	if err := service.HandleRequest("conn-1", addReq); err != nil {
		t.Fatalf("HandleRequest(add) error = %v", err)
	}

	addResp := collector.popLast(t)
	if addResp.Status != StatusOK || addResp.Action != ActionCronAdd {
		t.Fatalf("add response = %#v", addResp)
	}
	var addData CronAddResponse
	if err := json.Unmarshal(addResp.Data, &addData); err != nil {
		t.Fatalf("Unmarshal(add response) error = %v", err)
	}
	if addData.JobID == "" || addData.Job == nil {
		t.Fatalf("add data = %#v, want job_id and job", addData)
	}
	jobID := addData.JobID

	listReq := Request{
		Action: ActionCronList,
		Data:   rawJSON(t, CronListRequest{IncludeDisabled: true}),
	}
	if err := service.HandleRequest("conn-1", listReq); err != nil {
		t.Fatalf("HandleRequest(list) error = %v", err)
	}
	listResp := collector.popLast(t)
	if listResp.Status != StatusOK || listResp.Action != ActionCronList {
		t.Fatalf("list response = %#v", listResp)
	}
	var listData CronListResponse
	if err := json.Unmarshal(listResp.Data, &listData); err != nil {
		t.Fatalf("Unmarshal(list response) error = %v", err)
	}
	if len(listData.Jobs) != 1 {
		t.Fatalf("jobs len = %d, want 1", len(listData.Jobs))
	}
	if listData.Jobs[0].Description != "10分钟后提醒我喝水" {
		t.Fatalf("description = %q", listData.Jobs[0].Description)
	}
	if listData.Jobs[0].ScheduleType != "at" {
		t.Fatalf("schedule_type = %q, want at", listData.Jobs[0].ScheduleType)
	}

	updateReq := Request{
		Action: ActionCronUpdate,
		Data: rawJSON(t, CronUpdateRequest{
			JobID: jobID,
			CronAddRequest: CronAddRequest{
				Name:         "喝水提醒2",
				Description:  "每5分钟提醒我喝水",
				ScheduleType: "every",
				EverySeconds: 300,
				Command:      "echo hydrate-2",
			},
		}),
	}
	if err := service.HandleRequest("conn-1", updateReq); err != nil {
		t.Fatalf("HandleRequest(update) error = %v", err)
	}
	updateResp := collector.popLast(t)
	if updateResp.Status != StatusOK || updateResp.Action != ActionCronUpdate {
		t.Fatalf("update response = %#v", updateResp)
	}
	var updateData struct {
		Job CronJobInfo `json:"job"`
	}
	if err := json.Unmarshal(updateResp.Data, &updateData); err != nil {
		t.Fatalf("Unmarshal(update response) error = %v", err)
	}
	if updateData.Job.Name != "喝水提醒2" {
		t.Fatalf("updated job name = %q", updateData.Job.Name)
	}
	if updateData.Job.ScheduleType != "every" {
		t.Fatalf("updated schedule_type = %q, want every", updateData.Job.ScheduleType)
	}
	if updateData.Job.EverySeconds == nil || *updateData.Job.EverySeconds != 300 {
		t.Fatalf("updated every_seconds = %#v, want 300", updateData.Job.EverySeconds)
	}

	disableReq := Request{
		Action: ActionCronDisable,
		Data:   rawJSON(t, CronEnableRequest{JobID: jobID, Enabled: false}),
	}
	if err := service.HandleRequest("conn-1", disableReq); err != nil {
		t.Fatalf("HandleRequest(disable) error = %v", err)
	}
	disableResp := collector.popLast(t)
	if disableResp.Status != StatusOK || disableResp.Action != ActionCronDisable {
		t.Fatalf("disable response = %#v", disableResp)
	}
	var disableData struct {
		JobID   string      `json:"job_id"`
		Enabled bool        `json:"enabled"`
		Job     CronJobInfo `json:"job"`
	}
	if err := json.Unmarshal(disableResp.Data, &disableData); err != nil {
		t.Fatalf("Unmarshal(disable response) error = %v", err)
	}
	if disableData.Enabled || disableData.Job.Enabled {
		t.Fatalf("disable data = %#v, want disabled job", disableData)
	}

	enableReq := Request{
		Action: ActionCronEnable,
		Data:   rawJSON(t, CronEnableRequest{JobID: jobID, Enabled: true}),
	}
	if err := service.HandleRequest("conn-1", enableReq); err != nil {
		t.Fatalf("HandleRequest(enable) error = %v", err)
	}
	enableResp := collector.popLast(t)
	if enableResp.Status != StatusOK || enableResp.Action != ActionCronEnable {
		t.Fatalf("enable response = %#v", enableResp)
	}
	var enableData struct {
		JobID   string      `json:"job_id"`
		Enabled bool        `json:"enabled"`
		Job     CronJobInfo `json:"job"`
	}
	if err := json.Unmarshal(enableResp.Data, &enableData); err != nil {
		t.Fatalf("Unmarshal(enable response) error = %v", err)
	}
	if !enableData.Enabled || !enableData.Job.Enabled {
		t.Fatalf("enable data = %#v, want enabled job", enableData)
	}

	removeReq := Request{
		Action: ActionCronRemove,
		Data:   rawJSON(t, CronRemoveRequest{JobID: jobID}),
	}
	if err := service.HandleRequest("conn-1", removeReq); err != nil {
		t.Fatalf("HandleRequest(remove) error = %v", err)
	}
	removeResp := collector.popLast(t)
	if removeResp.Status != StatusOK || removeResp.Action != ActionCronRemove {
		t.Fatalf("remove response = %#v", removeResp)
	}
	var removeData map[string]string
	if err := json.Unmarshal(removeResp.Data, &removeData); err != nil {
		t.Fatalf("Unmarshal(remove response) error = %v", err)
	}
	if removeData["job_id"] != jobID {
		t.Fatalf("remove job_id = %q, want %q", removeData["job_id"], jobID)
	}
}

func boolPtr(v bool) *bool {
	return &v
}
