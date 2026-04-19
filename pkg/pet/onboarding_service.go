package pet

import (
	"fmt"
	"strings"

	petconfig "github.com/sipeed/picoclaw/pkg/pet/config"
)

type OnboardingApplyResult struct {
	PetID           string `json:"pet_id"`
	Language        string `json:"language"`
	ReminderEnabled bool   `json:"reminder_enabled"`
	VoiceEnabled    bool   `json:"voice_enabled"`
}

func mapPersonalityToneToPersonaType(tone string) string {
	switch strings.TrimSpace(tone) {
	case "甜心夹子":
		return "gentle"
	case "阳光怪气":
		return "playful"
	case "抽象发癫":
		return "cool"
	default:
		return "gentle"
	}
}

func buildPetPersonaFromSnapshot(snapshot *petconfig.OnboardingSnapshot) string {
	if snapshot == nil || snapshot.Pet == nil {
		return ""
	}

	role := ""
	if snapshot.Profile != nil {
		role = strings.TrimSpace(snapshot.Profile.Role)
	}

	callUserAs := strings.TrimSpace(snapshot.Pet.FinalNickname)
	if callUserAs == "" {
		callUserAs = strings.TrimSpace(snapshot.Pet.Nickname)
	}
	if callUserAs == "" && snapshot.Profile != nil {
		callUserAs = strings.TrimSpace(snapshot.Profile.DisplayName)
	}
	if callUserAs == "" {
		callUserAs = "你"
	}

	parts := []string{
		fmt.Sprintf("说话风格偏向%s", strings.TrimSpace(snapshot.Pet.PersonalityTone)),
		fmt.Sprintf("活跃程度为%d", snapshot.Pet.ActivityLevel),
		fmt.Sprintf("对用户的称呼是%s", callUserAs),
	}
	if role != "" {
		parts = append(parts, fmt.Sprintf("熟悉用户主修方向%s", role))
	}

	return strings.Join(parts, "；") + "。"
}

func firstReminderInterval(snapshot *petconfig.OnboardingSnapshot) int {
	if snapshot == nil || snapshot.StudentInsights == nil || snapshot.StudentInsights.PressurePlan == nil {
		return 0
	}
	for _, interval := range snapshot.StudentInsights.PressurePlan.ReminderIntervalsMinutes {
		if interval > 0 {
			return interval
		}
	}
	return 0
}

func (s *PetService) ApplyOnboardingSnapshot(snapshot *petconfig.OnboardingSnapshot) (*OnboardingApplyResult, error) {
	if snapshot == nil {
		return nil, fmt.Errorf("onboarding snapshot is nil")
	}
	if s.configManager == nil {
		return nil, fmt.Errorf("config manager not available")
	}

	if err := s.configManager.SaveOnboardingSnapshot(snapshot); err != nil {
		return nil, err
	}

	appCfg := s.configManager.GetApp()
	if appCfg == nil {
		appCfg = petconfig.DefaultAppConfig()
	}

	if snapshot.Profile != nil && strings.TrimSpace(snapshot.Profile.Language) != "" {
		appCfg.Language = strings.TrimSpace(snapshot.Profile.Language)
	}

	effectiveReminder := true
	if snapshot.AppPreferences != nil {
		effectiveReminder = snapshot.AppPreferences.EnableDesktopBubble
	}
	if snapshot.Permissions != nil {
		effectiveReminder = effectiveReminder && snapshot.Permissions.PopupReminder
	}

	appCfg.ReminderEnabled = effectiveReminder
	appCfg.VoiceEnabled = effectiveReminder
	appCfg.ProactiveCare = effectiveReminder
	if interval := firstReminderInterval(snapshot); interval > 0 {
		appCfg.ProactiveIntervalMinutes = interval
	}

	s.configManager.SetAppConfig(appCfg)

	char := s.charManager.GetCurrent()
	if char != nil && snapshot.Pet != nil {
		persona := buildPetPersonaFromSnapshot(snapshot)
		personaType := mapPersonalityToneToPersonaType(snapshot.Pet.PersonalityTone)
		s.charManager.UpdateCharacter(char.ID, char.Name, persona, personaType)
	}

	if err := s.configManager.Save(); err != nil {
		return nil, err
	}

	result := &OnboardingApplyResult{
		PetID:           "",
		Language:        appCfg.Language,
		ReminderEnabled: appCfg.ReminderEnabled,
		VoiceEnabled:    appCfg.VoiceEnabled,
	}
	if char != nil {
		result.PetID = char.ID
	}

	return result, nil
}
