package emotion

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const stateFileName = "pet_emotion_state.json"

type persistedState struct {
	Emotions   SixEmotions     `json:"emotions"`
	MBTI       MBTIPersonality `json:"mbti"`
	Volatility float64         `json:"volatility"`
	LastUpdate int64           `json:"last_update"`
}

func (e *EmotionEngine) Load() error {
	if e.persistPath == "" {
		return nil
	}

	data, err := os.ReadFile(filepath.Join(e.persistPath, stateFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var st persistedState
	if err := json.Unmarshal(data, &st); err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.emotions = SixEmotions{
		Joy:      clamp(st.Emotions.Joy),
		Anger:    clamp(st.Emotions.Anger),
		Sadness:  clamp(st.Emotions.Sadness),
		Disgust:  clamp(st.Emotions.Disgust),
		Surprise: clamp(st.Emotions.Surprise),
		Fear:     clamp(st.Emotions.Fear),
	}
	e.personality = MBTIPersonality{
		IE: clamp(st.MBTI.IE),
		SN: clamp(st.MBTI.SN),
		TF: clamp(st.MBTI.TF),
		JP: clamp(st.MBTI.JP),
	}
	if st.Volatility >= VolatilityMin && st.Volatility <= VolatilityMax {
		e.volatility = st.Volatility
	}
	if st.LastUpdate > 0 {
		e.lastUpdate = time.Unix(st.LastUpdate, 0)
	} else {
		e.lastUpdate = time.Now()
	}

	return nil
}

func (e *EmotionEngine) Save() error {
	if e.persistPath == "" {
		return nil
	}

	e.mu.RLock()
	st := persistedState{
		Emotions:   e.emotions,
		MBTI:       e.personality,
		Volatility: e.volatility,
		LastUpdate: e.lastUpdate.Unix(),
	}
	e.mu.RUnlock()

	if err := os.MkdirAll(e.persistPath, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(e.persistPath, stateFileName), data, 0o644)
}
