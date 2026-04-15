const { spawn } = require('child_process')

const VITE_URL = 'http://localhost:5173'
const MAX_WAIT = 30000
const CHECK_INTERVAL = 500

async function waitForVite() {
  const startTime = Date.now()

  while (Date.now() - startTime < MAX_WAIT) {
    try {
      const response = await fetch(VITE_URL, { method: 'HEAD' })
      if (response.ok) {
        console.log('[Electron] Vite server is ready!')
        return true
      }
    } catch {
      // Not ready yet
    }
    await new Promise(resolve => setTimeout(resolve, CHECK_INTERVAL))
  }

  console.error('[Electron] Timeout waiting for Vite server')
  return false
}

async function main() {
  const ready = await waitForVite()
  if (!ready) {
    process.exit(1)
  }

  console.log('[Electron] Starting Electron...')

  const electron = spawn('npx', ['electron', 'src/main.js'], {
    stdio: 'inherit',
    shell: true,
    cwd: process.cwd()
  })

  electron.on('close', (code) => {
    process.exit(code || 0)
  })
}

main()