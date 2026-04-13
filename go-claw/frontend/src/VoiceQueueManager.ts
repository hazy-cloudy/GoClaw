// 语音优先级
export const VOICE_PRIORITY = {
  REMINDER: 1,      // 提醒 - 最高优先级
  AI_REPLY: 2,      // AI回复 - 中优先级
  UI_INTERACT: 3,   // 点击/待机 - 最低优先级
} as const

export type VoicePriority = typeof VOICE_PRIORITY[keyof typeof VOICE_PRIORITY]

// 语音任务结构
interface VoiceTask {
  text: string
  priority: VoicePriority
  timestamp: number
}

// 语音队列管理器
export class VoiceQueueManager {
  private queue: VoiceTask[] = []
  private isPlaying: boolean = false
  private currentPriority: number = 0
  private onQueueChange?: (size: number) => void

  constructor(onQueueChange?: (size: number) => void) {
    this.onQueueChange = onQueueChange
  }

  // 添加语音到队列
  add(text: string, priority: VoicePriority): void {
    if (!text || text.trim() === '') return

    const task: VoiceTask = {
      text: text.trim(),
      priority,
      timestamp: Date.now()
    }

    // 插入到队列并排序（优先级高的在前）
    this.queue.push(task)
    this.queue.sort((a, b) => {
      if (a.priority !== b.priority) {
        return a.priority - b.priority
      }
      return a.timestamp - b.timestamp
    })

    this.onQueueChange?.(this.queue.length)
    this.processNext()
  }

  // 处理下一条语音
  private async processNext(): Promise<void> {
    // 如果正在播放，跳过
    if (this.isPlaying) return

    const task = this.queue.shift()
    if (!task) return

    this.isPlaying = true
    this.currentPriority = task.priority
    this.onQueueChange?.(this.queue.length)

    console.log(`[VoiceQueue] Playing: "${task.text}" (priority: ${task.priority})`)

    try {
      // 调用现有的 Speak 函数
      const { Speak } = await import('../wailsjs/go/main/TTSManager')
      await Speak(task.text)
    } catch (error) {
      console.error('[VoiceQueue] Speak error:', error)
    }

    this.isPlaying = false
    this.currentPriority = 0

    // 继续处理下一条
    this.processNext()
  }

  // 清空队列
  clear(): void {
    this.queue = []
    this.onQueueChange?.(0)
  }

  // 获取队列长度
  size(): number {
    return this.queue.length
  }

  // 是否正在播放
  isBusy(): boolean {
    return this.isPlaying
  }

  // 获取当前播放的优先级
  getCurrentPriority(): number {
    return this.currentPriority
  }

  // 跳过当前播放（可选）
  skip(): void {
    // 实际上无法中断已经在播放的语音
    // 但可以清空队列
    this.clear()
  }
}
