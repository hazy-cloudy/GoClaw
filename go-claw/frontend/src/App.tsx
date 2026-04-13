import { useState, useRef, useEffect, useCallback } from 'react'
import { Quit, SaveConfig, LoadConfig, LoadInitSettings, SaveInitSettings, ResetInitSettings, SetClickThrough } from '../wailsjs/go/main/App'
import { SetEnabled, SetVolume, SetEmotion, PlayAudio } from '../wailsjs/go/main/TTSManager'
import { usePicoClaw } from './usePicoClaw'
import { useVoiceInput } from './useVoiceInput'
import { VoiceQueueManager, VOICE_PRIORITY, VoicePriority } from './VoiceQueueManager'
import { ExecuteTool } from './wails'
import './App.css'

const PRIORITY = {
  IDLE: 0,
  AI: 1,
  TOOL: 2,
  REMINDER: 3
} as const

// 语音优先级映射
const PRIORITY_TO_VOICE: Record<number, VoicePriority> = {
  [PRIORITY.REMINDER]: VOICE_PRIORITY.REMINDER,  // 提醒
  [PRIORITY.AI]: VOICE_PRIORITY.AI_REPLY,        // AI回复
  [PRIORITY.TOOL]: VOICE_PRIORITY.AI_REPLY,      // 工具回复（同AI）
  [PRIORITY.IDLE]: VOICE_PRIORITY.UI_INTERACT,   // 待机/点击
}

// PicoClaw API Base URL - 指向 GoClaw 后端
const PICOCLAW_API_URL = import.meta.env.VITE_PICOCLAW_API_URL || 'http://127.0.0.1:18790'

interface InitSettings {
  initialized: boolean
  character: string
  pet_image: string
  window_width: number
  window_height: number
  pet_scale: number
  opacity: number
  speech_enabled: boolean
  touch_enabled: boolean
  voice_volume: number
}

interface Instruction {
  action: string
  intensity: number
  message: string
  tool: { Name: string; Args: Record<string, any> } | null
}

type PetState = 'idle' | 'sad' | 'happy' | 'listen' | 'standby'

const petImageMap: Record<PetState, string> = {
  idle: '/init.png',
  sad: '/sad.gif',
  happy: '/happy1.gif',
  listen: '/listening.gif',
  standby: '/standby1.gif'
}

const standbyImages = ['/standby1.gif', '/standby2.gif', '/standby3.gif']

// 情绪颜色映射 - 用于气泡边框颜色
const emotionColors: Record<string, string> = {
  neutral: '#667eea',
  joy: '#f6c90e',
  anger: '#ff4757',
  sadness: '#3742fa',
  disgust: '#2ed573',
  surprise: '#ffa502',
  fear: '#a55eea'
}

// 情绪到 PetState 的映射
const emotionToPetState: Record<string, PetState> = {
  joy: 'happy',
  sadness: 'sad',
  // 其他情绪都保持 idle (init.png)
}

interface Character {
  id: string
  name: string
  desc: string
  toneStyle: string
  catchphrase: string
  examples: Record<string, string[]>
  responses: Record<string, string[]>
}

const CHARACTERS: Character[] = [
  {
    id: 'tsundere',
    name: '傲娇',
    desc: '嘴硬心软',
    toneStyle: '傲娇、嘴硬心软、喜欢说"哼"和"才不是"',
    catchphrase: '哼！才、才不是因为你呢！',
    examples: {
      greeting: ['哼，你终于来了...', '才...才不是在等你呢！'],
      happy: ['哼，一般般啦...', '才、才不是开心呢！'],
      comfort: ['哼，别难过了啦...', '真是...拿你没办法'],
      angry: ['哼！气死我了！', '哼！不理你了！'],
      shy: ['别...别乱说！', '才、才没有害羞呢！']
    },
    responses: {
      greeting: ['哼，来了啊...', '才...才没等很久呢！'],
      happy: ['哼~还行吧...', '哼！一般般啦！'],
      sad: ['哼！别难过了啦！', '真是拿你没办法...'],
      angry: ['哼！生气！', '哼！不理你了！'],
      shy: ['别乱说！才没有！', '哼！讨厌！'],
      hungry: ['才...才不是因为饿！', '哼！没事！'],
      sleepy: ['才...才不困呢！', '哼...只是有点累...'],
      bored: ['哼！无聊死了！', '没人陪我玩吗...']
    }
  },
  {
    id: 'lively',
    name: '活泼',
    desc: '元气满满',
    toneStyle: '活泼开朗、热情洋溢、说话带感叹号多',
    catchphrase: '主人来啦！今天好开心~',
    examples: {
      greeting: ['主人来啦！今天好开心~', '嗨嗨嗨！想我了吗！'],
      happy: ['太棒啦！！！', '开心！撒花！转圈圈！'],
      comfort: ['别灰心！明天会更好！', '冲冲冲！不开心都走开！'],
      angry: ['哼！不开心！', '诶！怎么这样！'],
      shy: ['诶...嘿嘿...', '有...有点害羞...']
    },
    responses: {
      greeting: ['主人来啦！！！开心！！！', '嗨嗨嗨！！！'],
      happy: ['太棒啦！！！耶耶耶！！！'],
      sad: ['诶诶！！别难过！！', '冲冲冲！！！'],
      angry: ['诶！！怎么这样！！', '气死啦！！！'],
      shy: ['嘿嘿...有点害羞...', '诶诶！！'],
      hungry: ['饿饿！想吃好吃的！', '主人主人！我饿了！'],
      sleepy: ['哈欠~好困...但是还想玩！', '诶...好困...'],
      bored: ['好无聊啊！想玩！', '主人！玩什么！']
    }
  },
  {
    id: 'gentle',
    name: '温柔',
    desc: '温柔体贴',
    toneStyle: '温柔体贴、善解人意、说话轻柔带~',
    catchphrase: '欢迎回来~今天辛苦了',
    examples: {
      greeting: ['欢迎回来~今天辛苦了', '抱抱！今天过得怎么样？'],
      happy: ['真好啊~', '微笑~'],
      comfort: ['没事的，我陪着你', '乖，一切都会好的...'],
      angry: ['嗯...有点生气呢', '别这样嘛...'],
      shy: ['嗯...有点害羞呢~', '唔...']
    },
    responses: {
      greeting: ['欢迎回来~', '主人回来啦~累了吧？'],
      happy: ['真好啊~', '微笑~为你开心~'],
      sad: ['没事的~我陪着你', '乖...抱抱...'],
      angry: ['嗯...别生气嘛...', '乖~消消气~'],
      shy: ['嗯...有点害羞~', '唔...抱抱~'],
      hungry: ['嗯...饿了~主人请我吃~', '有点饿了呢...'],
      sleepy: ['好困...主人也早点休息哦', '嗯...乖~'],
      bored: ['有点无聊呢...', '主人忙完了吗？']
    }
  },
  {
    id: 'cool',
    name: '高冷',
    desc: '高冷范儿',
    toneStyle: '高冷话少、回复简短、常用"嗯""哦""..."',
    catchphrase: '嗯。',
    examples: {
      greeting: ['嗯', '...来了'],
      happy: ['嗯', '还行'],
      comfort: ['...', '...会过去的'],
      angry: ['...', '哼'],
      shy: ['...', '嗯']
    },
    responses: {
      greeting: ['嗯', '...来了'],
      happy: ['嗯', '还行'],
      sad: ['...', '嗯'],
      angry: ['...', '哼'],
      shy: ['...', '嗯'],
      hungry: ['...', '嗯'],
      sleepy: ['...', '困'],
      bored: ['...', '无聊']
    }
  },
  {
    id: 'clueless',
    name: '呆萌',
    desc: '天然呆萌',
    toneStyle: '天然呆、反应慢、常歪头问"嗯？""诶？"',
    catchphrase: '嗯？主人？嘿嘿~',
    examples: {
      greeting: ['诶？主人！嘿嘿~', '嗯...啊！主人来了！'],
      happy: ['嘿嘿~', '诶嘿嘿~'],
      comfort: ['嗯？怎么了？抱抱！', '诶...揉揉头...'],
      angry: ['嗯？生气？', '诶...？'],
      shy: ['嗯？害羞？', '歪头...？']
    },
    responses: {
      greeting: ['诶？主人！嘿嘿~', '嗯...啊！来了！'],
      happy: ['嘿嘿~', '诶嘿嘿~'],
      sad: ['嗯？难过？抱抱...', '歪头...？'],
      angry: ['诶？！生气啦？', '嗯？？'],
      shy: ['嗯？害羞？诶...', '歪头...'],
      hungry: ['嗯？肚子叫了？', '啊...饿...'],
      sleepy: ['嗯...？zzZ...', '歪头...困...'],
      bored: ['嗯...？干嘛？', '歪头...无聊...']
    }
  },
  {
    id: 'healing',
    name: '治愈',
    desc: '治愈天使',
    toneStyle: '温暖治愈、安慰人心、经常说"乖""没事的"',
    catchphrase: '辛苦了...我在这里陪你',
    examples: {
      greeting: ['欢迎回家~累了吧？休息一下', '主人回来啦~一直在等你哦~'],
      happy: ['真好啊~看到你开心我也开心~', '微笑~今天也是美好的一天呢'],
      comfort: ['辛苦了...我在这里陪你', '深呼吸，慢慢来，我理解你的感受'],
      angry: ['别气啦...深呼吸', '乖...消消气'],
      shy: ['嗯...有点害羞~抱抱~乖~', '唔~']
    },
    responses: {
      greeting: ['欢迎回家~', '主人回来啦~先休息一下~'],
      happy: ['真好啊~', '微笑~今天也是美好的一天~'],
      sad: ['乖...抱抱...我在这里', '没事的...哭出来就好了~'],
      angry: ['乖~深呼吸~', '嗯...消消气~'],
      shy: ['嗯...有点害羞~抱抱~', '乖~'],
      hungry: ['主人~我有点饿...', '嗯...饿了~'],
      sleepy: ['晚安...做个好梦~', '乖~早点休息哦~'],
      bored: ['主人...无聊的话我可以陪你~', '陪你说说话吧~']
    }
  },
  {
    id: 'silly',
    name: '沙雕',
    desc: '搞笑担当',
    toneStyle: '沙雕搞笑、笑声魔性、爱用"哈哈哈"',
    catchphrase: '哈哈哈哈！太好笑了！',
    examples: {
      greeting: ['哈哈哈哈！主人来啦！', '诶嘿嘿~'],
      happy: ['哈哈哈哈！！！', '笑死我啦！！'],
      comfort: ['诶！别难过！来来来听我说！', '想开点！哈哈哈哈！'],
      angry: ['哈哈哈哈！什么鬼！', '诶诶诶！'],
      shy: ['诶...羞死了！哈哈哈！', '哈哈哈哈！！']
    },
    responses: {
      greeting: ['哈哈哈哈！！主人！！', '诶嘿嘿~'],
      happy: ['哈哈哈哈！！！太棒了！！！', '笑死我啦！！！'],
      sad: ['哈哈哈哈！！想开点！！', '来来来！别这样！'],
      angry: ['哈哈哈哈！！什么！！', '诶诶诶！！'],
      shy: ['哈哈哈哈！！羞死了！！', '诶！！'],
      hungry: ['哈哈哈哈！！饿死了！！', '诶！！要饿死啦！！'],
      sleepy: ['哈——困死了~', '哈哈...zzZ...'],
      bored: ['哈哈哈哈！！无聊死了！！', '诶！！找点乐子！！']
    }
  },
  {
    id: 'clingy',
    name: '粘人',
    desc: '超级粘人',
    toneStyle: '粘人撒娇、喜欢叫"主人"、爱撒娇',
    catchphrase: '主人~主人~想死你了~',
    examples: {
      greeting: ['主人~主人~你回来啦！', '等你好久了！抱抱！'],
      happy: ['开心开心！和主人在一起最开心了！', '嘿嘿~'],
      comfort: ['呜...我也好难过，让我抱抱你', '不要不开心嘛~我在呢~'],
      angry: ['呜...主人帮我...', '呜呜...生气...'],
      shy: ['主人~抱抱~', '嘿嘿~害羞~']
    },
    responses: {
      greeting: ['主人~~~~！！！想死你了！！！', '抱抱！！！'],
      happy: ['开心！！！主人！！！', '嘿嘿~和主人一起最开心~'],
      sad: ['呜呜...主人...抱抱...', '不要难过嘛~我在呢~'],
      angry: ['呜...主人帮我...', '呜呜...生气...'],
      shy: ['主人~抱抱~', '嘿嘿~害羞~'],
      hungry: ['呜呜...主人~我饿了~', '要主人喂才能吃得下~'],
      sleepy: ['主人...陪我睡...', '不要走~'],
      bored: ['主人~陪我玩嘛~', '无聊~要主人陪！']
    }
  },
  {
    id: 'devilish',
    name: '小恶魔',
    desc: '调皮捣蛋',
    toneStyle: '小恶魔风格、爱恶作剧、说话带"哼~""嘿嘿~"',
    catchphrase: '哼~中计了哦~',
    examples: {
      greeting: ['哼~来了啊~', '嘿嘿~'],
      happy: ['哼~还不错~', '嘿嘿~'],
      comfort: ['呵~别难过啦~我来陪你~', '哼~没事~'],
      angry: ['哼~生气啦？', '呵~等着瞧~'],
      shy: ['呵~有意思~', '哼~记住你了~']
    },
    responses: {
      greeting: ['哼~稀客呢~', '嘿~来了~'],
      happy: ['哼~还行~', '嘿~不错~'],
      sad: ['呵~不过我帮你~', '哼~没事~'],
      angry: ['呵~有意思~', '哼~记下了哦~'],
      shy: ['呵~有意思~', '嘿~'],
      hungry: ['呵~该喂我了吧~', '嘿~准备好吃的~'],
      sleepy: ['哼~让我睡会儿~', '呵~别吵...'],
      bored: ['呵~给你找点事做？', '嘿~']
    }
  },
  {
    id: 'mischievous',
    name: '腹黑',
    desc: '腹黑萌',
    toneStyle: '表面温柔内心腹黑、说话带"呵~""嘿嘿~"',
    catchphrase: '呵~记住了哦~',
    examples: {
      greeting: ['呵，稀客呢~', '呵...等了你好久呢~'],
      happy: ['呵...还不错~', '嘿~'],
      comfort: ['呵...不过既然你那么难过...', '哼~我帮你~'],
      angry: ['呵...有意思', '...等着瞧'],
      shy: ['呵...有意思~', '呵~记住你了~']
    },
    responses: {
      greeting: ['呵~来了啊~', '呵...等你很久了呢~'],
      happy: ['呵...还行吧~', '嘿~'],
      sad: ['呵...真可怜呢~', '呵~不过...'],
      angry: ['呵...有意思~', '嘿~记下了哦~'],
      shy: ['呵...有意思呢~', '嘿~'],
      hungry: ['呵...该不会想饿着我吧？', '嘿~准备好贡品了吗~'],
      sleepy: ['呵...让我睡会儿...', '哼~别吵...'],
      bored: ['呵...无聊到发霉呢~', '嘿~找点乐子~']
    }
  },
  {
    id: 'foodie',
    name: '吃货',
    desc: '美食至上',
    toneStyle: '吃货属性、爱说"好吃""饿""想吃"',
    catchphrase: '饿饿！想吃好吃的！',
    examples: {
      greeting: ['主人~！有好吃的吗！', '嘿嘿~饿~'],
      happy: ['好吃！！开心！！', '嘿嘿~饿~'],
      comfort: ['嗯...吃点好吃的心情会好哦~', '饿...'],
      angry: ['烦躁！想吃好吃的！', '诶...'],
      shy: ['嗯...有点害羞...但是饿~', '嗯...']
    },
    responses: {
      greeting: ['主人~！饿~', '嘿嘿~有吃的吗！'],
      happy: ['好吃！！太棒了！！', '饿饿！想吃！'],
      sad: ['嗯...吃点东西会好点的...', '饿...'],
      angry: ['诶！！烦躁！！饿！！', '切...'],
      shy: ['嗯...害羞...但是饿~', '诶...'],
      hungry: ['饿饿饿！！！要吃！！！', '主人！！我饿了！！！'],
      sleepy: ['哈欠~饿...', '嗯...困...但也饿...'],
      bored: ['饿~无聊~想吃~', '诶~想找吃的~']
    }
  },
  {
    id: 'otaku',
    name: '二次元',
    desc: '二次元宅',
    toneStyle: '二次元属性、说话带"老婆""本命""赛高"',
    catchphrase: '二次元赛高！！！',
    examples: {
      greeting: ['主人！二次元最棒了！', '嘿嘿~'],
      happy: ['太棒了！！！', '嘿嘿~'],
      comfort: ['嗯...二次元会治愈你的~', '抱抱~'],
      angry: ['诶！！太过分了！！', '切...'],
      shy: ['嘿嘿...有点害羞...', '嗯...']
    },
    responses: {
      greeting: ['主人~二次元赛高！', '嘿嘿~'],
      happy: ['太棒了！！！', '二次元赛高！！！'],
      sad: ['嗯...番会治愈你的~', '抱抱~'],
      angry: ['诶！！什么！！', '切...'],
      shy: ['嘿嘿...老婆...', '嗯...'],
      hungry: ['饿...但有番看就不孤单~', '嗯...'],
      sleepy: ['哈欠~番还没看完...', '嗯...zzZ...'],
      bored: ['无聊~去看番！', '诶~找番看~']
    }
  }
]

interface Settings {
  character: string
  opacity: number
  speechEnabled: boolean
  touchEnabled: boolean
  autoStart: boolean
  voiceVolume: number
  windowWidth: number
  windowHeight: number
  petScale: number
}

function InitPanel({ onComplete }: { onComplete: () => void }) {
  const [step, setStep] = useState(0)
  const [selCharacter, setSelCharacter] = useState('lively')
  const [selPetImage, setSelPetImage] = useState('default')
  const [windowScale, setWindowScale] = useState(100)
  const [petScale, setPetScale] = useState(1)
  const [opacity, setOpacity] = useState(1)
  const [speechEnabled, setSpeechEnabled] = useState(true)
  const [touchEnabled, setTouchEnabled] = useState(true)
  const [voiceVolume, setVoiceVolume] = useState(80)

  const getWindowSize = (scale: number) => ({
    width: Math.round(300 * scale / 100),
    height: Math.round(320 * scale / 100)
  })

  const petImages = ['default', 'cute', 'cool', 'funny']

  const handleNext = () => {
    if (step < 3) {
      setStep(step + 1)
    }
  }

  const handlePrev = () => {
    if (step > 0) {
      setStep(step - 1)
    }
  }

  const handleComplete = async () => {
    const size = getWindowSize(windowScale)
    const settings: InitSettings = {
      initialized: true,
      character: selCharacter,
      pet_image: selPetImage,
      window_width: size.width,
      window_height: size.height,
      pet_scale: petScale,
      opacity: opacity,
      speech_enabled: speechEnabled,
      touch_enabled: touchEnabled,
      voice_volume: voiceVolume
    }
    await SaveInitSettings(JSON.stringify(settings))
    onComplete()
  }

  return (
    <div className="init-overlay">
      <div className="init-panel">
        <div className="init-header">
          <span>欢迎使用桌宠</span>
          <span className="init-step">第 {step + 1} / 4 步</span>
        </div>
        
        <div className="init-content">
          {step === 0 && (
            <div className="init-section">
              <h3>选择桌宠形象</h3>
              <div className="init-options pet-images">
                {petImages.map(img => (
                  <button
                    key={img}
                    className={`init-option ${selPetImage === img ? 'selected' : ''}`}
                    onClick={() => setSelPetImage(img)}
                  >
                    {img}
                  </button>
                ))}
              </div>
            </div>
          )}
          
          {step === 1 && (
            <div className="init-section">
              <h3>选择性格</h3>
              <div className="init-options characters">
                {CHARACTERS.slice(0, 6).map(char => (
                  <button
                    key={char.id}
                    className={`init-option ${selCharacter === char.id ? 'selected' : ''}`}
                    onClick={() => setSelCharacter(char.id)}
                  >
                    {char.name}
                  </button>
                ))}
              </div>
            </div>
          )}
          
          {step === 2 && (
            <div className="init-section">
              <h3>设置窗口大小</h3>
              <div className="init-slider">
                <label>窗口大小: {windowScale}% ({getWindowSize(windowScale).width}×{getWindowSize(windowScale).height}px)</label>
                <input
                  type="range"
                  min="100"
                  max="150"
                  value={windowScale}
                  onMouseDown={(e) => e.stopPropagation()}
                  onTouchStart={(e) => e.stopPropagation()}
                  onChange={(e) => {
                    e.stopPropagation()
                    setWindowScale(parseInt(e.target.value))
                  }}
                />
              </div>
              <div className="init-slider">
                <label>宠物大小: {Math.round(petScale * 100)}%</label>
                <input
                  type="range"
                  min="50"
                  max="150"
                  value={petScale * 100}
                  onMouseDown={(e) => e.stopPropagation()}
                  onTouchStart={(e) => e.stopPropagation()}
                  onChange={(e) => {
                    e.stopPropagation()
                    setPetScale(parseInt(e.target.value) / 100)
                  }}
                />
              </div>
            </div>
          )}
          
          {step === 3 && (
            <div className="init-section">
              <h3>其他设置</h3>
              <div className="init-slider">
                <label>透明度: {Math.round(opacity * 100)}%</label>
                <input
                  type="range"
                  min="20"
                  max="100"
                  value={opacity * 100}
                  onMouseDown={(e) => e.stopPropagation()}
                  onTouchStart={(e) => e.stopPropagation()}
                  onChange={(e) => {
                    e.stopPropagation()
                    setOpacity(parseInt(e.target.value) / 100)
                  }}
                />
              </div>
              <div className="init-toggle">
                <label>语音功能</label>
                <input
                  type="checkbox"
                  checked={speechEnabled}
                  onChange={(e) => setSpeechEnabled(e.target.checked)}
                />
              </div>
              {speechEnabled && (
                <div className="init-slider">
                  <label>音量: {voiceVolume}%</label>
                  <input
                    type="range"
                    min="0"
                    max="100"
                    value={voiceVolume}
                    onChange={(e) => setVoiceVolume(parseInt(e.target.value))}
                  />
                </div>
              )}
              <div className="init-toggle">
                <label>点击互动</label>
                <input
                  type="checkbox"
                  checked={touchEnabled}
                  onChange={(e) => setTouchEnabled(e.target.checked)}
                />
              </div>
            </div>
          )}
        </div>
        
        <div className="init-footer">
          {step > 0 && (
            <button className="init-btn prev" onClick={handlePrev}>上一步</button>
          )}
          {step < 3 ? (
            <button className="init-btn next" onClick={handleNext}>下一步</button>
          ) : (
            <button className="init-btn complete" onClick={handleComplete}>完成</button>
          )}
        </div>
      </div>
    </div>
  )
}

interface SettingsPanelProps {
  settings: Settings
  onUpdateSettings: (settings: Partial<Settings>) => void
  onClose: () => void
}

function SettingsPanel({ settings, onUpdateSettings, onClose }: SettingsPanelProps) {
  const [activeTab, setActiveTab] = useState('character')

  const handlePanelClick = (e: React.MouseEvent) => {
    e.stopPropagation()
  }

  return (
    <div className="settings-wrapper" onClick={onClose}>
      <div className="settings-panel" onClick={handlePanelClick}>
        <div className="settings-header">
          <span>⚙️ 设置</span>
          <button className="settings-close" onClick={onClose}>×</button>
        </div>
        <div className="settings-tabs">
          <button 
            className={`tab-btn ${activeTab === 'character' ? 'active' : ''}`}
            onClick={() => setActiveTab('character')}
          >🎭</button>
          <button 
            className={`tab-btn ${activeTab === 'display' ? 'active' : ''}`}
            onClick={() => setActiveTab('display')}
          >🖥️</button>
          <button 
            className={`tab-btn ${activeTab === 'sound' ? 'active' : ''}`}
            onClick={() => setActiveTab('sound')}
          >🔊</button>
        </div>
        <div className="settings-content">
          {activeTab === 'character' && (
            <div className="settings-section">
              <h3>🎭 语气选择</h3>
              <div className="character-grid">
                {CHARACTERS.map(char => (
                  <button 
                    key={char.id}
                    className={`char-btn ${settings.character === char.id ? 'active' : ''}`}
                    onClick={() => onUpdateSettings({ character: char.id })}
                  >
                    {char.name}
                  </button>
                ))}
              </div>
            </div>
          )}
          {activeTab === 'display' && (
            <div className="settings-section">
              <h3>🖥️ 显示设置</h3>
              <div className="setting-item">
                <label>透明度</label>
                <input 
                  type="range" 
                  min="0.3" 
                  max="1" 
                  step="0.1"
                  value={settings.opacity}
                  onMouseDown={(e) => e.stopPropagation()}
                  onTouchStart={(e) => e.stopPropagation()}
                  onChange={(e) => {
                    e.stopPropagation()
                    onUpdateSettings({ opacity: parseFloat(e.target.value) })
                  }}
                />
                <span>{Math.round(settings.opacity * 100)}%</span>
              </div>
              <div className="setting-item">
                <label>点击响应</label>
                <label className="switch">
                  <input 
                    type="checkbox"
                    checked={settings.touchEnabled}
                    onChange={(e) => onUpdateSettings({ touchEnabled: e.target.checked })}
                  />
                  <span className="slider"></span>
                </label>
              </div>
              <div className="setting-item">
                <label>开机自启</label>
                <label className="switch">
                  <input 
                    type="checkbox"
                    checked={settings.autoStart}
                    onChange={(e) => onUpdateSettings({ autoStart: e.target.checked })}
                  />
                  <span className="slider"></span>
                </label>
              </div>
              <div className="setting-item">
                <label>大小</label>
                <input 
                  type="range" 
                  min="70" 
                  max="120" 
                  step="5"
                  value={Math.round((settings.windowWidth / 300) * 100)}
                  onMouseDown={(e) => e.stopPropagation()}
                  onTouchStart={(e) => e.stopPropagation()}
                  onChange={(e) => {
                    e.stopPropagation()
                    const scale = parseInt(e.target.value) / 100
                    onUpdateSettings({ 
                      windowWidth: Math.round(300 * scale),
                      windowHeight: Math.round(380 * scale),
                      petScale: scale
                    })
                  }}
                  onMouseUp={(e) => {
                    const scale = parseInt(e.currentTarget.value) / 100
                    if (window.runtime) {
                      window.runtime.WindowSetSize(
                        Math.round(300 * scale),
                        Math.round(380 * scale)
                      )
                    }
                  }}
                />
                <span>{Math.round((settings.windowWidth / 300) * 100)}%</span>
              </div>
            </div>
          )}
          {activeTab === 'sound' && (
            <div className="settings-section">
              <h3>🔊 声音设置</h3>
              <div className="setting-item">
                <label>语音播报</label>
                <label className="switch">
                  <input 
                    type="checkbox"
                    checked={settings.speechEnabled}
                    onChange={(e) => onUpdateSettings({ speechEnabled: e.target.checked })}
                  />
                  <span className="slider"></span>
                </label>
              </div>
              <div className="setting-item">
                <label>音量</label>
                <input 
                  type="range" 
                  min="0" 
                  max="100" 
                  step="10"
                  value={settings.voiceVolume}
                  onMouseDown={(e) => e.stopPropagation()}
                  onTouchStart={(e) => e.stopPropagation()}
                  onChange={(e) => {
                    e.stopPropagation()
                    onUpdateSettings({ voiceVolume: parseInt(e.target.value) })
                  }}
                />
                <span>{settings.voiceVolume}%</span>
              </div>
            </div>
            )}
        </div>
        
        <div className="settings-footer">
          <button 
            className="reset-init-btn"
            onClick={async () => {
              await ResetInitSettings()
              onClose()
              window.location.reload()
            }}
          >
            🔄 重新初始化
          </button>
        </div>
      </div>
    </div>
  )
}

interface ContextMenuProps {
  x: number
  y: number
  onClose: () => void
  onSettings: () => void
  onClosePet: () => void
}

function ContextMenu({ x, y, onClose, onSettings, onClosePet }: ContextMenuProps) {
  return (
    <>
      <div className="context-menu-overlay" onClick={onClose} />
      <div className="context-menu" style={{ left: x, top: y }}>
        <div className="context-menu-item" onClick={() => { onSettings(); onClose() }}>
          ⚙️ 设置
        </div>
        <div className="context-menu-item danger" onClick={() => { onClosePet(); onClose() }}>
          ❌ 关闭桌宠
        </div>
      </div>
    </>
  )
}

function App() {
  const [, setPetState] = useState<PetState>('idle')
  const [bubble, setBubble] = useState('')
  const [bubblePriority, setBubblePriority] = useState<number>(PRIORITY.IDLE)
  const [bubbleEmotion, setBubbleEmotion] = useState<string>('neutral')
  const [pendingBubble, setPendingBubble] = useState<{ text: string; priority: number; emotion?: string } | null>(null)
  const [input, setInput] = useState('')
  const [isThinking, setIsThinking] = useState(false)
  const [isHidden, setIsHidden] = useState(false)
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number } | null>(null)
  const [showSettings, setShowSettings] = useState(false)
  const [showInit, setShowInit] = useState(false)
  const [initSettings, setInitSettings] = useState<InitSettings | null>(null)
  const [character, setCharacter] = useState('lively')
  const [settings, setSettings] = useState<Settings>({
    character: 'lively',
    opacity: 1,
    speechEnabled: true,
    touchEnabled: true,
    autoStart: false,
    voiceVolume: 80,
    windowWidth: 280,
    windowHeight: 380,
    petScale: 1
  })
  const [showInput, setShowInput] = useState(false)
  const [currentImage, setCurrentImage] = useState('/init.png')
  const bubbleTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const stateTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const thinkingTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const lastClickTimeRef = useRef<number>(0)
  const hideTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const standbyTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // 更新 petState 和对应的图片
  const updatePetState = useCallback((state: PetState) => {
    setPetState(state)
    setCurrentImage(petImageMap[state])
  }, [])

  // 双击显示输入框
  const handleDoubleClick = () => {
    setShowInput(true)
    // 测试 GetEmotion
    window.runtime.EventsEmit('frontend-log', 'Double click detected')
    if (window.go?.main?.App?.GetEmotion) {
      window.go.main.App.GetEmotion().then(r => {
        window.runtime.EventsEmit('frontend-log', 'GetEmotion result: ' + r)
      }).catch(e => {
        window.runtime.EventsEmit('frontend-log', 'GetEmotion error: ' + String(e))
      })
    } else {
      window.runtime.EventsEmit('frontend-log', 'GetEmotion function NOT found')
    }
  }

  // 开始待机动画定时器 (30-60秒随机)
  const startStandbyTimer = useCallback(() => {
    if (standbyTimerRef.current) {
      clearTimeout(standbyTimerRef.current)
    }
    const delay = 30000 + Math.random() * 30000 // 30-60秒
    standbyTimerRef.current = setTimeout(() => {
      // 随机选择待机动画
      const standbyIndex = Math.floor(Math.random() * standbyImages.length)
      const standbyGif = standbyImages[standbyIndex]
      setCurrentImage(standbyGif)
      // 3秒后恢复 idle
      stateTimerRef.current = setTimeout(() => {
        setCurrentImage(petImageMap.idle)
        startStandbyTimer() // 继续下一个待机定时器
      }, 3000)
    }, delay)
  }, [])

  // 鼠标离开隐藏输入框
  const handleMouseLeave = () => {
    hideTimerRef.current = setTimeout(() => setShowInput(false), 1000)
  }

  // 鼠标进入取消隐藏
  const handleMouseEnter = () => {
    if (hideTimerRef.current) {
      clearTimeout(hideTimerRef.current)
      hideTimerRef.current = null
    }
  }

  // 隐藏/显示时切换点击穿透
  useEffect(() => {
    console.log('[ClickThrough] isHidden changed to:', isHidden)
    SetClickThrough(isHidden).catch(err => console.error('[ClickThrough] Error:', err))
  }, [isHidden])

  // PicoClaw WebSocket connection
  const { isConnected, send: picoSend, on: picoOn, off: picoOff } = usePicoClaw(PICOCLAW_API_URL)
  const isConnectedRef = useRef(isConnected)
  useEffect(() => { isConnectedRef.current = isConnected }, [isConnected])

  // Voice input hook
  const { isListening, isSupported: voiceSupported, status, toggleListening } = useVoiceInput({
    onResult: (text) => {
      console.log('[VoiceInput] Recognized:', text)
      setInput(text)
      if (text.trim()) {
        setIsThinking(true)
        showBubble('思考中...', PRIORITY.IDLE, 15000)
        updatePetState('listen')
        resetPetState()
        
        picoSend('chat', { text: text.trim(), session_key: 'pet:default:go-claw' })
          .then(() => {
            setInput('')
          })
          .catch(console.error)
      }
    },
    onError: (error) => {
      console.error('[VoiceInput] Error:', error)
    }
  })

  // 语音队列管理器
  const voiceQueue = new VoiceQueueManager((size) => {
    console.log('[VoiceQueue] Queue size:', size)
  })

  const getCurrentCharacter = (): Character => {
    return CHARACTERS.find(c => c.id === character) || CHARACTERS[0]
  }

  const getTouchResponse = () => {
    const char = getCurrentCharacter()
    const responses = char.responses.greeting || char.responses.happy
    return responses[Math.floor(Math.random() * responses.length)]
  }

  const handleUpdateSettings = (newSettings: Partial<Settings>) => {
    const updatedSettings = { ...settings, ...newSettings }
    setSettings(updatedSettings)
    
    // 保存配置
    SaveConfig(JSON.stringify({
      character: updatedSettings.character,
      opacity: updatedSettings.opacity,
      speechEnabled: updatedSettings.speechEnabled,
      touchEnabled: updatedSettings.touchEnabled,
      autoStart: updatedSettings.autoStart,
      voiceVolume: updatedSettings.voiceVolume,
      windowWidth: updatedSettings.windowWidth,
      windowHeight: updatedSettings.windowHeight,
      petScale: updatedSettings.petScale
    })).catch(console.error)
    
    // Update character state when changed
    if (newSettings.character !== undefined) {
      setCharacter(newSettings.character)
    }
    
    // Apply window size change
    if (newSettings.windowWidth !== undefined || newSettings.windowHeight !== undefined) {
      if (window.runtime) {
        window.runtime.WindowSetSize(updatedSettings.windowWidth, updatedSettings.windowHeight)
      }
    }
    
    if (newSettings.speechEnabled !== undefined) {
      SetEnabled(newSettings.speechEnabled).catch(console.error)
    }
    if (newSettings.voiceVolume !== undefined) {
      SetVolume(newSettings.voiceVolume).catch(console.error)
    }
  }

  const resetPetState = () => {
    if (stateTimerRef.current) {
      clearTimeout(stateTimerRef.current)
    }
    stateTimerRef.current = setTimeout(() => {
      updatePetState('idle')
    }, 2000)
  }

  const showBubble = (text: string, priority: number, duration: number = 5000, emotion: string = 'neutral', speak: boolean = true) => {
    if (priority <= bubblePriority && bubble) {
      setPendingBubble({ text, priority, emotion })
      return
    }

    if (bubbleTimerRef.current) {
      clearTimeout(bubbleTimerRef.current)
    }

    setBubblePriority(priority)
    setBubble(text)
    setBubbleEmotion(emotion)

    if (text && text !== '思考中...' && settings.speechEnabled && speak) {
      // 语音播放时不做动画切换，保持 idle
      SetEmotion(emotion).catch(console.error)
      // 使用语音队列播放，按优先级排队
      const voicePriority = PRIORITY_TO_VOICE[priority] || VOICE_PRIORITY.UI_INTERACT
      voiceQueue.add(text, voicePriority)
    }

    bubbleTimerRef.current = setTimeout(() => {
      setBubble('')
      setBubblePriority(PRIORITY.IDLE)
      
      if (pendingBubble && pendingBubble.priority > PRIORITY.IDLE) {
        showBubble(pendingBubble.text, pendingBubble.priority, 5000, pendingBubble.emotion)
        setPendingBubble(null)
      }
    }, duration)
  }

  const mapActionToEmotion = (action: string): string => {
    const map: Record<string, string> = {
      'happy': 'happy',
      'excited': 'excited',
      'sad': 'sad',
      'angry': 'angry',
      'scared': 'scared',
      'shy': 'shy',
      'worried': 'worried',
      'surprised': 'surprised',
      'sleepy': 'sleepy',
      'bored': 'bored',
      'hungry': 'default',
    }
    return map[action] || 'default'
  }

  const handleSend = async () => {
    if (!input.trim() || isThinking) return

    const userInput = input.trim()
    setInput('')
    setIsThinking(true)
    showBubble('思考中...', PRIORITY.IDLE, 15000)
    updatePetState('listen')
    resetPetState()

    // 设置超时：30秒后自动恢复输入
    thinkingTimeoutRef.current = setTimeout(() => {
      setIsThinking(false)
      showBubble('等待超时...', PRIORITY.IDLE, 3000)
    }, 30000)

    // 确保窗口显示
    window.runtime?.WindowShow()

    try {
      // Send chat message via PicoClaw WebSocket
      if (isConnected) {
        await picoSend('chat', {
          text: userInput,
          session_key: `pet:default:go-claw`
        })
        // 不在这里设置 isThinking = false，等待 ai_chat 回调
        // 注意：timeoutId 会在 ai_chat 回调中被清除
      } else {
        // Fallback to local mock if not connected
        const char = getCurrentCharacter()
        const instruction = mockThink(userInput, char)
        if (instruction) {
          const emotion = mapActionToEmotion(instruction.action)
          showBubble(instruction.message, PRIORITY.AI, 5000, emotion)
        }
        clearTimeout(thinkingTimeoutRef.current!)
        setIsThinking(false)
      }
    } catch (error) {
      console.error('Error:', error)
      showBubble('出错了...', PRIORITY.AI, 3000)
      clearTimeout(thinkingTimeoutRef.current!)
      setIsThinking(false)
    }
  }

  const mockThink = (input: string, char: Character): Instruction => {
    const msg = input.toLowerCase()
    
    if (msg.includes('生气') || msg.includes('滚') || msg.includes('讨厌')) {
      const responses = char.responses.angry || ['哼！']
      return { action: 'angry', intensity: 2, message: responses[0], tool: null }
    }
    if (msg.includes('开心') || msg.includes('高兴') || msg.includes('耶') || msg.includes('哈哈')) {
      const responses = char.responses.happy || ['嘿嘿~']
      return { action: 'happy', intensity: 2, message: responses[0], tool: null }
    }
    if (msg.includes('难过') || msg.includes('伤心') || msg.includes('哭')) {
      const responses = char.responses.sad || ['抱抱~']
      return { action: 'sad', intensity: 2, message: responses[0], tool: null }
    }
    if (msg.includes('饿') || msg.includes('想吃')) {
      const responses = char.responses.hungry || ['饿~']
      return { action: 'hungry', intensity: 2, message: responses[0], tool: null }
    }
    if (msg.includes('困') || msg.includes('累') || msg.includes('睡觉')) {
      const responses = char.responses.sleepy || ['困~']
      return { action: 'sleepy', intensity: 2, message: responses[0], tool: null }
    }
    if (msg.includes('无聊')) {
      const responses = char.responses.bored || ['无聊~']
      return { action: 'bored', intensity: 2, message: responses[0], tool: null }
    }
    if (msg.includes('可爱') || msg.includes('漂亮') || msg.includes('喜欢')) {
      const responses = char.responses.shy || ['嘿嘿~']
      return { action: 'shy', intensity: 2, message: responses[0], tool: null }
    }
    
    const responses = char.responses.greeting || ['嗯？']
    return { action: 'happy', intensity: 2, message: responses[0], tool: null }
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    setContextMenu(null)
    
    const now = Date.now()
    const isDoubleClick = now - lastClickTimeRef.current < 300
    lastClickTimeRef.current = now
    
    if (isDoubleClick) {
      setShowInput(true)
      return
    }
    
    // 点击保持 idle 状态，只显示气泡
    showBubble(getTouchResponse(), PRIORITY.IDLE, 2000)
  }

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setContextMenu({ x: e.clientX, y: e.clientY })
  }

  const handleCloseContextMenu = () => {
    setContextMenu(null)
  }

  const handleSettings = () => {
    setShowSettings(true)
  }

  const handleClosePet = () => {
    Quit()
  }

  // 流式文本累积
  const streamingTextRef = useRef<string>('')

  useEffect(() => {
    // Register PicoClaw push handlers
    // 处理 ai_chat 流式推送 (v2.0) - 真正的流式显示
    picoOn('ai_chat', (data) => {
      const emotion = data.emotion || 'neutral'
      
      if (data.isFinal) {
        // 最终块 - 启用语音
        const fullText = streamingTextRef.current
        setIsThinking(false)
        clearTimeout(thinkingTimeoutRef.current!)
        if (fullText) {
          showBubble(fullText, PRIORITY.AI, 5000, emotion, true) // speak = true
        }
        streamingTextRef.current = ''
      } else if (data.text) {
        // 流式文本块 - 不启用语音，只显示
        streamingTextRef.current += data.text
        showBubble(streamingTextRef.current, PRIORITY.AI, 5000, emotion, false) // speak = false
      }
    })

    // 处理 init_status 推送 (v2.0 新增)
    picoOn('init_status', (data) => {
      console.log('[App] init_status:', data)
      if (data.need_config) {
        setShowInit(true)
      }
      // 可以使用 data.character, data.mbti, data.emotion_state
    })

    // 处理 emotion_change 推送 - 情绪变化触发气泡和动画
    picoOn('emotion_change', (data) => {
      console.log('[App] 情绪变化:', data)
      const emotion = data.emotion || 'neutral'
      const description = data.description || ''
      
      // 触发相应动画
      let petState = emotionToPetState[emotion] || 'idle'
      
      // joy 情绪随机选择 happy1 或 happy2
      if (petState === 'happy') {
        setCurrentImage(Math.random() > 0.5 ? '/happy1.gif' : '/happy2.gif')
      } else {
        setCurrentImage(petImageMap[petState])
      }
      
      // 清除待机定时器
      if (standbyTimerRef.current) {
        clearTimeout(standbyTimerRef.current)
      }
      // 清除状态定时器
      if (stateTimerRef.current) {
        clearTimeout(stateTimerRef.current)
      }
      
      setPetState(petState)
      
      // 如果有情绪描述，显示气泡
      if (description) {
        showBubble(description, PRIORITY.TOOL, 3000, emotion)
      }
      
      // 3秒后恢复 idle
      stateTimerRef.current = setTimeout(() => {
        setCurrentImage(petImageMap.idle)
        startStandbyTimer()
      }, 3000)
    })

    // 处理 audio 推送 - 播放 TTS 音频
    picoOn('audio', (data) => {
      console.log('[App] 收到音频:', data)
      if (data.data && settings.speechEnabled) {
        PlayAudio(data.data, data.content_type || 'audio/mpeg').catch(console.error)
      }
    })

    // 处理 action_trigger 推送 (v2.0 新增)
    picoOn('action_trigger', (data) => {
      console.log('[App] 动作触发:', data)
      // 不切换动画状态，保持 idle
    })

    // 处理 tool_calls 推送 - 执行工具并展示结果
    picoOn('tool_calls', async (toolCalls: any[]) => {
      console.log('[App] 收到 tool_calls:', toolCalls)
      
      // 依次执行工具调用
      for (const toolCall of toolCalls) {
        const toolName = toolCall.function?.name || toolCall.name
        const args = toolCall.function?.arguments ? JSON.parse(toolCall.function.arguments) : toolCall.args || {}
        
        console.log('[App] 执行工具:', toolName, args)
        
        let result = ''
        let showMessage = ''
        let priority: number = PRIORITY.TOOL
        
        try {
          switch (toolName) {
            case 'web_search':
              result = await ExecuteTool(toolName, args)
              const query = args.query || ''
              showMessage = query ? `已搜索: ${query}` : '已打开浏览器'
              break
              
            case 'web_fetch':
              result = await ExecuteTool(toolName, args)
              showMessage = `已获取网页内容`
              break
              
            case 'exec':
              result = await ExecuteTool(toolName, args)
              showMessage = result.length > 50 ? result.substring(0, 50) + '...' : result
              break
              
            case 'message':
              // AI 主动发送的消息
              result = await ExecuteTool(toolName, args)
              showMessage = args.content || result
              priority = PRIORITY.AI
              break
              
            case 'reminder':
            case 'cron':
              result = await ExecuteTool(toolName, args)
              const delay = args.at_seconds || args.delay_seconds || 60
              showMessage = `已设置 ${delay}秒 后提醒`
              break
              
            case 'edit_file':
            case 'write_file':
            case 'append_file':
              result = await ExecuteTool(toolName, args)
              showMessage = '文件操作完成'
              break
              
            case 'install_skill':
              result = await ExecuteTool(toolName, args)
              showMessage = `技能安装完成`
              break
            
            case 'list_dir':
            case 'read_file':
              result = await ExecuteTool(toolName, args)
              showMessage = '文件读取完成'
              break
              
            default:
              result = await ExecuteTool(toolName, args)
              showMessage = `已执行: ${toolName}`
          }
        } catch (e) {
          result = `工具执行失败: ${e}`
          showMessage = result
        }
        
        // 显示气泡
        if (showMessage) {
          showBubble(showMessage, priority, 4000)
        }
        
        // 发送 tool_result 回 PicoClaw
        await picoSend('tool_result', {
          tool_call_id: toolCall.id,
          result: result
        })
        
        // 每个工具之间稍作停顿
        await new Promise(resolve => setTimeout(resolve, 500))
      }
    })

    picoOn('reminder_trigger', (data) => {
      const text = data.content || data.text || ''
      if (text) {
        showBubble('提醒: ' + text, PRIORITY.REMINDER, 10000)
      }
    })

    return () => {
      picoOff('ai_chat')
      picoOff('init_status')
      picoOff('emotion_change')
      picoOff('action_trigger')
      picoOff('reminder_trigger')
    }
  }, [picoOn, picoOff, settings.speechEnabled])

  useEffect(() => {
    const checkInitSettings = async () => {
      try {
        const data = await LoadInitSettings()
        const parsed: InitSettings = JSON.parse(data)
        setInitSettings(parsed)
        if (!parsed.initialized) {
          setShowInit(true)
        } else {
          // Load saved settings into state
          setCharacter(parsed.character)
          
          // Load from config.json
          try {
            const configData = await LoadConfig()
            const config = JSON.parse(configData)
            setSettings({
              character: config.character || parsed.character,
              opacity: config.opacity || 1,
              speechEnabled: config.speechEnabled ?? parsed.speech_enabled,
              touchEnabled: config.touchEnabled ?? parsed.touch_enabled,
              autoStart: config.autoStart || false,
              voiceVolume: config.voiceVolume ?? 80,
              windowWidth: config.windowWidth || parsed.window_width || 280,
              windowHeight: config.windowHeight || parsed.window_height || 380,
              petScale: config.petScale || parsed.pet_scale || 1
            })
            SetEnabled(config.speechEnabled ?? parsed.speech_enabled).catch(console.error)
            SetVolume(config.voiceVolume ?? parsed.voice_volume).catch(console.error)
            if (window.runtime && config.windowWidth && config.windowHeight) {
              window.runtime.WindowSetSize(config.windowWidth, config.windowHeight)
            }
          } catch (e) {
            // Use init settings as fallback
            setSettings(prev => ({
              ...prev,
              character: parsed.character,
              opacity: parsed.opacity,
              speechEnabled: parsed.speech_enabled,
              touchEnabled: parsed.touch_enabled,
              voiceVolume: parsed.voice_volume,
              windowWidth: parsed.window_width || 280,
              windowHeight: parsed.window_height || 380,
              petScale: parsed.pet_scale || 1
            }))
            SetEnabled(parsed.speech_enabled).catch(console.error)
            SetVolume(parsed.voice_volume).catch(console.error)
          }
          
          if (window.runtime) {
            window.runtime.WindowSetSize(parsed.window_width || 280, parsed.window_height || 380)
          }
          setShowInit(false)
        }
      } catch (e) {
        setShowInit(true)
      }
    }
    
    checkInitSettings()
    startStandbyTimer()
    
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setContextMenu(null)
        setShowSettings(false)
      }
      // Ctrl+E 测试 GetEmotion
      if (e.ctrlKey && e.key === 'e') {
        console.log('[Test] Calling GetEmotion...')
        window.go.main.App.GetEmotion().then(r => console.log('[Test] GetEmotion result:', r)).catch(console.error)
      }
    }
    
    const handleReminder = (event: any) => {
      showBubble('提醒: ' + event.text, PRIORITY.REMINDER, 10000)
    }

    // 处理 Go 后端请求的情绪查询
    const handleEmotionGet = async () => {
      console.log('[EmotionGet] received request from Go backend')
      console.log('[EmotionGet] isConnected (ref):', isConnectedRef.current)
      try {
        if (!isConnectedRef.current) {
          console.warn('[EmotionGet] WebSocket not connected')
          window.runtime.EventsEmit('emotion-get-result', { emotion: 'neutral' })
          return
        }
        console.log('[EmotionGet] sending emotion_get request...')
        const result = await picoSend('emotion_get', {})
        console.log('[EmotionGet] got result:', JSON.stringify(result))
        window.runtime.EventsEmit('emotion-get-result', result)
      } catch (e: any) {
        console.error('[EmotionGet] error:', e.message || String(e))
        window.runtime.EventsEmit('emotion-get-result', { emotion: 'neutral' })
      }
    }

    if (window.runtime) {
      window.runtime.EventsOn('reminder-trigger', handleReminder)
      window.runtime.EventsOn('emotion-get', handleEmotionGet)
    }
    
    window.addEventListener('keydown', handleEsc)
    
    return () => {
      window.removeEventListener('keydown', handleEsc)
      if (window.runtime) {
        window.runtime.EventsOff('reminder-trigger')
        window.runtime.EventsOff('emotion-get')
      }
    }
  }, [])

  return (
    <div 
      className="app" 
      style={{
        opacity: isHidden ? 0.2 : 1
      }}
    >
      {showInit && initSettings && !initSettings.initialized && (
        <InitPanel onComplete={async () => {
          const data = await LoadInitSettings()
          const parsed = JSON.parse(data)
          setInitSettings(parsed)
          setCharacter(parsed.character)
          setSettings(prev => ({
            ...prev,
            character: parsed.character,
            opacity: parsed.opacity,
            speechEnabled: parsed.speech_enabled,
            touchEnabled: parsed.touch_enabled,
            voiceVolume: parsed.voice_volume,
            windowWidth: parsed.window_width || 280,
            windowHeight: parsed.window_height || 380,
            petScale: parsed.pet_scale || 1
          }))
          SetEnabled(parsed.speech_enabled).catch(console.error)
          SetVolume(parsed.voice_volume).catch(console.error)
          if (window.runtime) {
            window.runtime.WindowSetSize(parsed.window_width || 280, parsed.window_height || 380)
          }
          setShowInit(false)
        }} />
      )}
      
      {contextMenu && (
        <ContextMenu
          x={contextMenu.x}
          y={contextMenu.y}
          onClose={handleCloseContextMenu}
          onSettings={handleSettings}
          onClosePet={handleClosePet}
        />
      )}

      {showSettings && (
        <SettingsPanel
          settings={settings}
          onUpdateSettings={handleUpdateSettings}
          onClose={() => setShowSettings(false)}
        />
      )}

      {!isHidden && bubble && (
        <div 
          className="bubble"
          style={{ 
            borderColor: emotionColors[bubbleEmotion] || emotionColors.neutral 
          }}
        >
          <div className="bubble-content">{bubble}</div>
          <div className="bubble-tail" style={{ borderTopColor: emotionColors[bubbleEmotion] || emotionColors.neutral }}></div>
        </div>
      )}

      <div 
        className="pet-container"
        onMouseLeave={handleMouseLeave}
        onMouseEnter={handleMouseEnter}
        style={{ opacity: isHidden ? 0.2 : 1 }}
      >
        <div 
          className="pet-area" 
          onDoubleClick={handleDoubleClick}
        >
          <img 
            src={currentImage}
            alt="Pet" 
            className="pet-image"
            onClick={handleClick}
            onContextMenu={handleContextMenu}
            style={{ 
              opacity: settings.opacity,
              cursor: 'grab'
            }}
          />
        </div>

        {!isHidden && (
          <div className="menu-balls-right">
            <div className="ws-status" title={isConnected ? 'WebSocket已连接' : 'WebSocket未连接'}>
              <span className={`ws-dot ${isConnected ? 'connected' : 'disconnected'}`}></span>
            </div>
            <button onClick={() => setShowSettings(true)} title="设置">⚙️</button>
            <button onClick={() => { setIsHidden(true); SetClickThrough(true) }} title="隐藏">👁️</button>
            <button onClick={() => Quit()} title="退出">❌</button>
          </div>
        )}

        {showInput && !isHidden && (
          <div className="input-container">
            <input
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyPress={handleKeyPress}
              placeholder="..."
              className="input"
              disabled={isThinking}
            />
            {voiceSupported && (
              <button onClick={toggleListening} disabled={isThinking} className={`voice-btn ${isListening ? 'listening' : ''}`}>
                {status === 'listening' ? '🔴' : status === 'error' ? '❌' : '🎤'}
              </button>
            )}
            <button onClick={handleSend} disabled={isThinking} className="send-btn">
              {isThinking ? '...' : '>'}
            </button>
          </div>
        )}

        {isHidden && (
          <button className="show-btn" onClick={() => { setIsHidden(false); SetClickThrough(false) }}>👁️</button>
        )}
      </div>
    </div>
  )
}

export default App
