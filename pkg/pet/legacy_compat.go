package pet

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/pkg/pet/config"
)

type AppConfig = config.AppConfig

type Character struct {
	PetID       string    `json:"pet_id"`
	Name        string    `json:"name"`
	Persona     string    `json:"persona"`
	PersonaType string    `json:"persona_type"`
	Avatar      string    `json:"avatar"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CharacterStore struct {
	path string
	mu   sync.RWMutex
	char *Character
}

func GetCharacterStoreWithPath(workspacePath string) *CharacterStore {
	return &CharacterStore{path: filepath.Join(workspacePath, "pet_character.json")}
}

func (s *CharacterStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.char = defaultCharacter()
			return nil
		}
		return err
	}

	var c Character
	if err := json.Unmarshal(data, &c); err != nil {
		return err
	}

	if c.PetID == "" {
		c.PetID = "pet_001"
	}
	if c.Avatar == "" {
		c.Avatar = "default"
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}

	s.char = &c
	return nil
}

func (s *CharacterStore) IsOnboardingComplete() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.char == nil {
		return false
	}
	return s.char.Name != "" && s.char.Persona != "" && s.char.PersonaType != ""
}

func (s *CharacterStore) Get() *Character {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.char == nil {
		return defaultCharacter()
	}
	clone := *s.char
	return &clone
}

func (s *CharacterStore) Update(input *Character) error {
	if input == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if s.char == nil {
		s.char = defaultCharacter()
	}

	if input.PetID != "" {
		s.char.PetID = input.PetID
	}
	if input.Name != "" {
		s.char.Name = input.Name
	}
	if input.Persona != "" {
		s.char.Persona = input.Persona
	}
	if input.PersonaType != "" {
		s.char.PersonaType = input.PersonaType
	}
	if input.Avatar != "" {
		s.char.Avatar = input.Avatar
	}
	if s.char.CreatedAt.IsZero() {
		s.char.CreatedAt = now
	}
	s.char.UpdatedAt = now

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.char, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func defaultCharacter() *Character {
	now := time.Now()
	return &Character{
		PetID:       "pet_001",
		Name:        "",
		Persona:     "",
		PersonaType: "",
		Avatar:      "default",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
