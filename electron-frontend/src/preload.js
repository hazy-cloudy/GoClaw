const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  openDashboard: () => ipcRenderer.send('open-dashboard'),
  openOnboarding: () => ipcRenderer.send('open-onboarding'),
  setOnboardingMode: (enabled) => ipcRenderer.send('set-onboarding-mode', enabled),
  getBackendBaseUrl: () => process.env.GOCLAW_BACKEND_URL || 'http://127.0.0.1:18800',
  minimizeWindow: () => ipcRenderer.send('window-minimize'),
  toggleMaximizeWindow: () => ipcRenderer.send('window-toggle-maximize'),
  closeWindow: () => ipcRenderer.send('window-close'),
  showBubble: (text, emotion, audio) => {
    ipcRenderer.send('show-bubble', { text, emotion, audio })
  },
  sendConnectionAlive: () => ipcRenderer.send('connection-alive'),
  setClickThrough: (enabled) => ipcRenderer.send('set-click-through', enabled),
  onBubbleShow: (callback) => {
    ipcRenderer.on('bubble-show', (event, data) => callback(data));
  },
  onConnectionAlive: (callback) => {
    ipcRenderer.on('connection-alive', () => callback());
  }
});
