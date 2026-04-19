package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const OnboardingSnapshotFile = "pet_onboarding.json"

type OnboardingSnapshot struct {
	Version          int                         `json:"version"`
	Completed        bool                        `json:"completed"`
	CompletedAt      string                      `json:"completed_at"`
	Profile          *OnboardingProfile          `json:"profile,omitempty"`
	Pet              *OnboardingPet             `json:"pet,omitempty"`
	AppPreferences   *OnboardingAppPreferences   `json:"app_preferences,omitempty"`
	Permissions      *OnboardingPermissions      `json:"permissions,omitempty"`
	StudyPreferences *OnboardingStudyPreferences `json:"study_preferences,omitempty"`
	StudentInsights  *OnboardingStudentInsights  `json:"student_insights,omitempty"`
}

type OnboardingProfile struct {
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	Language    string `json:"language"`
}

type OnboardingPet struct {
	PersonalityTone string `json:"personality_tone"`
	ActivityLevel   int    `json:"activity_level"`
	Nickname        string `json:"nickname"`
	CustomNickname  string `json:"custom_nickname"`
	FinalNickname   string `json:"final_nickname"`
	VoiceStyle      string `json:"voice_style"`
}

type OnboardingAppPreferences struct {
	AutoConnectOnLaunch   bool `json:"auto_connect_on_launch"`
	EnableDesktopBubble   bool `json:"enable_desktop_bubble"`
	OpenConsoleOnPetClick bool `json:"open_console_on_pet_click"`
}

type OnboardingPermissions struct {
	ScreenView    bool `json:"screen_view"`
	AppCheck      bool `json:"app_check"`
	LocalDocRead  bool `json:"local_doc_read"`
	PopupReminder bool `json:"popup_reminder"`
}

type OnboardingStudyPreferences struct {
	SleepHour        int      `json:"sleep_hour"`
	AnxietyLevel     int      `json:"anxiety_level"`
	SelectedBreakers []string `json:"selected_breakers"`
	IcsFileName      string   `json:"ics_file_name"`
}

type OnboardingLearningRhythm struct {
	Chronotype      string   `json:"chronotype"`
	FocusWindows    []string `json:"focus_windows"`
	QuietWindows    []string `json:"quiet_windows"`
	ReminderCadence string   `json:"reminder_cadence"`
	Summary         string   `json:"summary"`
}

type OnboardingPressurePlan struct {
	Level                    string `json:"level"`
	Strategy                 string `json:"strategy"`
	ReminderIntervalsMinutes []int  `json:"reminder_intervals_minutes"`
	ToneGuide                string `json:"tone_guide"`
	TemplateSoft             string `json:"template_soft"`
	TemplateNormal           string `json:"template_normal"`
	TemplateStrong           string `json:"template_strong"`
}

type OnboardingStudentInsights struct {
	LearningRhythm *OnboardingLearningRhythm `json:"learning_rhythm,omitempty"`
	PressurePlan   *OnboardingPressurePlan   `json:"pressure_plan,omitempty"`
}

func (l *ConfigLoader) onboardingSnapshotPath() string {
	return filepath.Join(l.workspacePath, OnboardingSnapshotFile)
}

func (l *ConfigLoader) SaveOnboardingSnapshot(snapshot *OnboardingSnapshot) error {
	if snapshot == nil {
		return fmt.Errorf("onboarding snapshot is nil")
	}

	if err := os.MkdirAll(l.workspacePath, 0o755); err != nil {
		return fmt.Errorf("failed to create workspace dir: %w", err)
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal onboarding snapshot: %w", err)
	}

	if err := os.WriteFile(l.onboardingSnapshotPath(), data, 0o644); err != nil {
		return fmt.Errorf("failed to write onboarding snapshot: %w", err)
	}

	return nil
}

func (l *ConfigLoader) LoadOnboardingSnapshot() (*OnboardingSnapshot, error) {
	data, err := os.ReadFile(l.onboardingSnapshotPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read onboarding snapshot: %w", err)
	}

	var snapshot OnboardingSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to parse onboarding snapshot: %w", err)
	}

	return &snapshot, nil
}
