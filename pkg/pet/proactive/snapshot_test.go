package proactive

import (
	"testing"
	"time"

	"github.com/sipeed/picoclaw/pkg/pet/activity"
	"github.com/sipeed/picoclaw/pkg/pet/characters"
	petconfig "github.com/sipeed/picoclaw/pkg/pet/config"
	"github.com/sipeed/picoclaw/pkg/pet/userprofile"
)

func TestBuildSnapshotUsesUserTone(t *testing.T) {
	baseDir := t.TempDir()

	cfgMgr := petconfig.NewManager(baseDir)
	if cfgMgr == nil {
		t.Fatal("expected config manager")
	}

	charMgr, err := characters.NewManager(cfgMgr.GetCharacters(), cfgMgr)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	userMgr := userprofile.NewManager(baseDir, nil, charMgr, nil, "")
	userMgr.UpdateProfile(&userprofile.UserProfileUpdateRequest{
		PersonalityTone: "阴阳怪气",
		PressureLevel:   "high",
	})

	activityStore, err := activity.NewStore(baseDir)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	snapshot := BuildSnapshot(time.Now(), SnapshotDependencies{
		ActivityStore:      activityStore,
		ConfigManager:      cfgMgr,
		UserProfileManager: userMgr,
		CharacterProvider:  charMgr,
	})

	if snapshot.Pet.PersonalityTone != "阴阳怪气" {
		t.Fatalf("Pet.PersonalityTone = %q, want %q", snapshot.Pet.PersonalityTone, "阴阳怪气")
	}
	if snapshot.Pet.PersonaType != "gentle" {
		t.Fatalf("Pet.PersonaType = %q, want %q", snapshot.Pet.PersonaType, "gentle")
	}
}
