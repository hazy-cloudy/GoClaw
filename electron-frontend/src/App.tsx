import { useState, useEffect, useCallback, useRef } from 'react'

type PetState = 'idle' | 'sad' | 'happy' | 'listen' | 'standby'

const happyImages = ['/happy1.gif', '/happy2.gif']
const standbyImages = ['/standby1.gif', '/standby2.gif', '/standby3.gif']

const getPetImage = (state: PetState, prevImage?: string): string => {
  switch (state) {
    case 'idle': return '/init.png'
    case 'sad': return '/sad.gif'
    case 'happy': return happyImages[Math.floor(Math.random() * happyImages.length)]
    case 'listen': return '/listening.gif'
    case 'standby':
      let img = standbyImages[Math.floor(Math.random() * standbyImages.length)]
      if (prevImage && standbyImages.includes(prevImage) && standbyImages.length > 1) {
        while (img === prevImage) {
          img = standbyImages[Math.floor(Math.random() * standbyImages.length)]
        }
      }
      return img
    default: return '/init.png'
  }
}

const emotionToPetState: Record<string, PetState> = {
  joy: 'happy',
  happy: 'happy',
  sadness: 'sad',
  sad: 'sad',
  neutral: 'standby',
}

function App() {
  const [petState, setPetState] = useState<PetState>('idle')
  const [currentImage, setCurrentImage] = useState('/init.png')
  const currentImageRef = useRef(currentImage)
  currentImageRef.current = currentImage

  const [bubble, setBubble] = useState('')
  const audioRef = useRef<HTMLAudioElement | null>(null)
  const bubbleTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const transitionTo = useCallback((newState: PetState, delay?: number) => {
    const prevImage = currentImageRef.current
    setPetState(newState)
    setCurrentImage(getPetImage(newState, prevImage))
    if (delay) {
      setTimeout(() => {
        setPetState('standby')
        setCurrentImage(getPetImage('standby', currentImageRef.current))
      }, delay)
    }
  }, [])

  useEffect(() => {
    const handleBubbleShow = (data: { text: string | null; emotion?: string; audio?: string }) => {
      if (bubbleTimerRef.current) {
        clearTimeout(bubbleTimerRef.current)
        bubbleTimerRef.current = null
      }

      if (data.text !== null) {
        setBubble(data.text)
      }

      const targetState = data.emotion && emotionToPetState[data.emotion]
        ? emotionToPetState[data.emotion]
        : 'happy'
      transitionTo(targetState)

      if (data.audio) {
        let audioUrl = data.audio
        if (!audioUrl.startsWith('data:') && !audioUrl.startsWith('http')) {
          audioUrl = `data:audio/mp3;base64,${audioUrl}`
        }
        if (audioRef.current) {
          audioRef.current.pause()
        }
        audioRef.current = new Audio(audioUrl)
        audioRef.current.onended = () => {
          setBubble('')
          transitionTo('standby')
        }
        audioRef.current.play().catch(() => {
          setBubble('')
        })
      } else {
        bubbleTimerRef.current = setTimeout(() => {
          setBubble('')
          transitionTo('standby')
        }, 10000)
      }
    }

    window.electronAPI?.onBubbleShow(handleBubbleShow)
  }, [transitionTo])

  const openDashboard = () => {
    window.electronAPI?.openDashboard()
  }

  return (
    <div className="app">
      <div className="pet-container">
        <div className="controls">
          <button className="btn" onClick={openDashboard} title="控制台">💬</button>
          <button className="btn close" onClick={() => window.close()} title="关闭">×</button>
        </div>

        {bubble && <div className="bubble">{bubble}</div>}

        <div className="pet-area">
          <img className="pet-image" src={currentImage} alt="Pet" />
        </div>
      </div>
    </div>
  )
}

export default App
