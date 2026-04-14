package voice

import (
	"fmt"
	"os"
	"strings"

	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/pet/config"
)

// Loader 语音服务加载器
// 根据配置初始化TTS提供者
type Loader struct {
	cfg      *config.VoiceConfig // 语音配置
	provider StreamingTTS        // TTS提供者实例
}

// NewLoader 创建语音加载器实例
func NewLoader(cfg *config.VoiceConfig) *Loader {
	return &Loader{cfg: cfg}
}

// Load 根据配置加载并初始化TTS提供者
// 会根据DefaultModel或第一个启用的模型来确定使用哪个TTS
func (l *Loader) Load() error {
	if l.cfg == nil || len(l.cfg.ModelList) == 0 {
		logger.Warnf("pet voice: no voice models configured")
		return nil
	}

	// 确定要加载的模型名称：优先使用默认模型，否则选择第一个启用的
	modelName := l.cfg.DefaultModel
	if modelName == "" {
		for _, m := range l.cfg.ModelList {
			if m.Enabled {
				modelName = m.Name
				break
			}
		}
	}

	if modelName == "" {
		logger.Warnf("pet voice: no enabled voice model found")
		return nil
	}

	// 查找模型配置
	var modelCfg *config.VoiceModelConfig
	for _, m := range l.cfg.ModelList {
		if m.Name == modelName {
			modelCfg = m
			break
		}
	}

	if modelCfg == nil {
		return fmt.Errorf("voice model %s not found in model_list", modelName)
	}

	if !modelCfg.Enabled {
		return fmt.Errorf("voice model %s is disabled", modelName)
	}

	// 创建TTS提供者（目前仅支持Minimax）
	l.provider = newMinimaxTTS(modelCfg.APIBase, modelCfg.APIKey, modelCfg.Model)
	logger.Infof("pet voice: loaded TTS provider, model=%s", modelCfg.Model)
	return nil
}

// GetProvider 返回已加载的TTS提供者
func (l *Loader) GetProvider() StreamingTTS {
	return l.provider
}

// IsASREnabled 检查ASR（语音识别）功能是否启用
func (l *Loader) IsASREnabled() bool {
	return l.cfg != nil && l.cfg.ASREnabled
}

// resolveEnvVar 解析配置值中的环境变量引用
// 格式: ${ENV_VAR_NAME}，会替换为对应环境变量的值
func resolveEnvVar(value string) string {
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		envKey := value[2 : len(value)-1]
		return os.Getenv(envKey)
	}
	return value
}
