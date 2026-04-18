const { app, BrowserWindow, ipcMain, screen } = require('electron');
const path = require('path');
const os = require('os');
const fs = require('fs');

let petWindow = null;
let settingsWindow = null;

const PET_WIDTH = 280;
const PET_HEIGHT = 380;
const ONBOARDING_WIDTH = 1280;
const ONBOARDING_HEIGHT = 860;

const userDataPath = path.join(os.homedir(), '.goclaw');
app.setPath('userData', userDataPath);

if (!fs.existsSync(userDataPath)) {
  fs.mkdirSync(userDataPath, { recursive: true });
}

const logFilePath = path.join(userDataPath, 'logs.txt');
const logStream = fs.createWriteStream(logFilePath, { flags: 'a' });

function logToFile(message) {
  const timestamp = new Date().toISOString();
  const logLine = `[${timestamp}] ${message}\n`;
  logStream.write(logLine);
  console.log(logLine);
}

logToFile('Electron application started');

const rendererBaseUrl = (process.env.ELECTRON_RENDERER_URL || 'http://localhost:5173').trim().replace(/\/+$/, '');
const shouldOpenDevTools = process.env.ELECTRON_OPEN_DEVTOOLS === '1';

function createPetWindow() {
  const { width, height } = screen.getPrimaryDisplay().workAreaSize;

  petWindow = new BrowserWindow({
    width: PET_WIDTH,
    height: PET_HEIGHT,
    x: width - PET_WIDTH - 20,
    y: height - PET_HEIGHT - 60,
    frame: false,
    transparent: true,
    alwaysOnTop: true,
    resizable: false,
    skipTaskbar: true,
    show: false,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js')
    }
  });

  petWindow.loadURL(rendererBaseUrl).catch((err) => {
    logToFile(`[PET WINDOW] loadURL failed: ${String(err)}`);
  });

  petWindow.once('ready-to-show', () => {
    petWindow.show();
  });

  petWindow.on('closed', () => {
    petWindow = null;
    app.quit();
  });
}

function applyOnboardingMode(enabled) {
  if (!petWindow || petWindow.isDestroyed()) {
    return;
  }

  const display = screen.getDisplayMatching(petWindow.getBounds());
  const area = display.workArea;

  if (enabled) {
    const width = Math.min(ONBOARDING_WIDTH, area.width - 40);
    const height = Math.min(ONBOARDING_HEIGHT, area.height - 40);
    const x = area.x + Math.round((area.width - width) / 2);
    const y = area.y + Math.round((area.height - height) / 2);

    petWindow.setResizable(true);
    petWindow.setAlwaysOnTop(false);
    petWindow.setSkipTaskbar(false);
    petWindow.setBounds({ x, y, width, height }, true);
    return;
  }

  const x = area.x + area.width - PET_WIDTH - 20;
  const y = area.y + area.height - PET_HEIGHT - 60;
  petWindow.setResizable(false);
  petWindow.setAlwaysOnTop(true);
  petWindow.setSkipTaskbar(true);
  petWindow.setBounds({ x, y, width: PET_WIDTH, height: PET_HEIGHT }, true);
}

function createSettingsWindow() {
  if (settingsWindow && !settingsWindow.isDestroyed()) {
    if (settingsWindow.isMinimized()) {
      settingsWindow.restore();
    }
    settingsWindow.show();
    settingsWindow.focus();
    return;
  }

  const { width, height } = screen.getPrimaryDisplay().workAreaSize;
  const settingsWidth = Math.round(width * 0.7);
  const settingsHeight = Math.round(height * 0.7);

  settingsWindow = new BrowserWindow({
    width: settingsWidth,
    height: settingsHeight,
    minWidth: 600,
    minHeight: 400,
    show: false,
    frame: false,
    transparent: false,
    backgroundColor: '#111111',
    alwaysOnTop: false,
    autoHideMenuBar: true,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js')
    }
  });

  const settingsUrl = `${rendererBaseUrl}/settings.html`;
  logToFile(`[SETTINGS WINDOW] opening ${settingsUrl}`);

  settingsWindow.webContents.on('did-fail-load', (_event, code, desc, url) => {
    logToFile(`[SETTINGS WINDOW] did-fail-load code=${code} desc=${desc} url=${url}`);
  });

  settingsWindow.webContents.on('did-finish-load', () => {
    logToFile('[SETTINGS WINDOW] did-finish-load');
  });

  settingsWindow.loadURL(settingsUrl).catch((err) => {
    logToFile(`[SETTINGS WINDOW] loadURL failed: ${String(err)}`);
  });

  settingsWindow.once('ready-to-show', () => {
    settingsWindow.show();
    settingsWindow.focus();
    if (shouldOpenDevTools) {
      settingsWindow.webContents.openDevTools({ mode: 'detach' });
    }
  });

  settingsWindow.on('closed', () => {
    settingsWindow = null;
  });
}

ipcMain.on('open-settings', () => {
  logToFile('[IPC] open-settings');
  createSettingsWindow();
});

ipcMain.on('set-onboarding-mode', (_event, enabled) => {
  logToFile(`[IPC] set-onboarding-mode ${Boolean(enabled)}`);
  applyOnboardingMode(Boolean(enabled));
});

ipcMain.on('window-minimize', (event) => {
  const target = BrowserWindow.fromWebContents(event.sender);
  if (target && !target.isDestroyed()) {
    target.minimize();
  }
});

ipcMain.on('window-toggle-maximize', (event) => {
  const target = BrowserWindow.fromWebContents(event.sender);
  if (!target || target.isDestroyed()) {
    return;
  }
  if (target.isMaximized()) {
    target.unmaximize();
  } else {
    target.maximize();
  }
});

ipcMain.on('window-close', (event) => {
  const target = BrowserWindow.fromWebContents(event.sender);
  if (target && !target.isDestroyed()) {
    target.close();
  }
});

ipcMain.on('renderer-log', (_event, { level, args }) => {
  const message = args.map((arg) => {
    if (typeof arg === 'object') {
      return JSON.stringify(arg);
    }
    return String(arg);
  }).join(' ');

  if (level === 'error') {
    logToFile(`[RENDERER ERROR] ${message}`);
  } else {
    logToFile(`[RENDERER] ${message}`);
  }
});

ipcMain.on('minimize-settings', () => {
  if (settingsWindow && !settingsWindow.isDestroyed()) {
    settingsWindow.minimize();
  }
});

ipcMain.on('maximize-settings', () => {
  if (settingsWindow && !settingsWindow.isDestroyed()) {
    if (settingsWindow.isMaximized()) {
      settingsWindow.unmaximize();
    } else {
      settingsWindow.maximize();
    }
  }
});

ipcMain.on('close-settings', () => {
  if (settingsWindow && !settingsWindow.isDestroyed()) {
    settingsWindow.close();
  }
});

ipcMain.on('settings-changed', (_event, settings) => {
  if (petWindow && !petWindow.isDestroyed()) {
    petWindow.webContents.send('settings-updated', settings);
  }
});

ipcMain.on('chat-history', (_event, history) => {
  if (settingsWindow && !settingsWindow.isDestroyed()) {
    settingsWindow.webContents.send('chat-history-updated', history);
  }
});

ipcMain.on('show-bubble', (_event, data) => {
  if (petWindow && !petWindow.isDestroyed()) {
    petWindow.webContents.send('bubble-show', data);
  }
});

ipcMain.on('connection-alive', () => {
  if (petWindow && !petWindow.isDestroyed()) {
    petWindow.webContents.send('connection-alive');
  }
});

app.whenReady().then(() => {
  createPetWindow();
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit();
});
