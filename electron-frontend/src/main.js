const { app, BrowserWindow, ipcMain, screen } = require('electron');
const path = require('path');
const os = require('os');
const fs = require('fs');

let petWindow = null;
let settingsWindow = null;

// 璁剧疆鐢ㄦ埛鏁版嵁鐩綍浠ヨВ鍐虫潈闄愰棶棰?
const userDataPath = path.join(os.homedir(), '.goclaw');
app.setPath('userData', userDataPath);

if (!fs.existsSync(userDataPath)) {
  fs.mkdirSync(userDataPath, { recursive: true });
}

// 鍒涘缓鏃ュ織鏂囦欢
const logFilePath = path.join(userDataPath, 'logs.txt');
const logStream = fs.createWriteStream(logFilePath, { flags: 'a' });

function logToFile(message) {
  const timestamp = new Date().toISOString();
  const logLine = `[${timestamp}] ${message}\n`;
  logStream.write(logLine);
  console.log(logLine);
}

logToFile('Electron application started');

const rendererBaseUrl = (process.env.ELECTRON_RENDERER_URL || 'http://localhost:5173')
  .trim()
  .replace(/\/+$/, '');
const shouldOpenDevTools = process.env.ELECTRON_OPEN_DEVTOOLS === '1';

function revealWindow(win) {
  if (!win || win.isDestroyed()) {
    return;
  }
  if (win.isMinimized()) {
    win.restore();
  }
  win.show();
  win.moveTop();
  win.focus();
}

function createPetWindow() {
  // 鑾峰彇灞忓箷鍙充笅瑙掍綅锟?
  const { width, height } = screen.getPrimaryDisplay().workAreaSize;
  
  // 妗屽疇绐楀彛澶у皬
  const petWidth = 280;
  const petHeight = 380;
  
  petWindow = new BrowserWindow({
    width: petWidth,
    height: petHeight,
    x: width - petWidth - 20,
    y: height - petHeight - 60,
    frame: false,           // 鏃犺竟锟?
    transparent: true,      // 閫忔槑鑳屾櫙
    alwaysOnTop: true,     // 缃《
    resizable: false,       // 涓嶅彲璋冩暣澶у皬
    skipTaskbar: true,     // 涓嶆樉绀哄湪浠诲姟锟?
    show: false,           // 鍔犺浇瀹屾垚鍚庢樉锟?
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

function createSettingsWindow() {
  if (settingsWindow && !settingsWindow.isDestroyed()) {
    logToFile('[SETTINGS WINDOW] reusing existing window');
    revealWindow(settingsWindow);
    return;
  }

  const { width, height } = screen.getPrimaryDisplay().workAreaSize;
  const settingsWidth = Math.round(width * 0.7);
  const settingsHeight = Math.round(height * 0.7);
  const settingsUrl = `${rendererBaseUrl}/settings.html`;
  
  settingsWindow = new BrowserWindow({
    width: settingsWidth,
    height: settingsHeight,
    minWidth: 600,
    minHeight: 400,
    show: false,
    frame: false,
    transparent: false,
    backgroundColor: '#1e1e2e',
    alwaysOnTop: false,
    autoHideMenuBar: true,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js')
    }
  });

  let hasBeenRevealed = false;
  const revealSettingsWindow = () => {
    if (!settingsWindow || settingsWindow.isDestroyed() || hasBeenRevealed) {
      return;
    }
    hasBeenRevealed = true;
    logToFile('[SETTINGS WINDOW] reveal');
    revealWindow(settingsWindow);
    if (shouldOpenDevTools) {
      settingsWindow.webContents.openDevTools({ mode: 'detach' });
    }
  };

  logToFile(`[SETTINGS WINDOW] opening ${settingsUrl}`);

  settingsWindow.once('ready-to-show', revealSettingsWindow);
  settingsWindow.webContents.once('did-finish-load', () => {
    logToFile('[SETTINGS WINDOW] did-finish-load');
    revealSettingsWindow();
  });
  settingsWindow.webContents.on('console-message', (_event, level, message, line, sourceId) => {
    logToFile(`[SETTINGS WINDOW CONSOLE] level=${level} source=${sourceId}:${line} message=${message}`);
  });
  settingsWindow.webContents.on('did-fail-load', (_event, code, desc, url) => {
    logToFile(`[SETTINGS WINDOW] did-fail-load code=${code} desc=${desc} url=${url}`);
  });
  settingsWindow.webContents.on('render-process-gone', (_event, details) => {
    logToFile(`[SETTINGS WINDOW] render-process-gone reason=${details.reason} exitCode=${details.exitCode}`);
  });

  settingsWindow.loadURL(settingsUrl).catch((error) => {
    logToFile(`[SETTINGS WINDOW] loadURL failed: ${error}`);
  });

  settingsWindow.on('closed', () => {
    logToFile('[SETTINGS WINDOW] closed');
    settingsWindow = null;
  });
}

ipcMain.on('open-settings', () => {
  logToFile('[IPC] open-settings');
  createSettingsWindow();
});

// 鎺ユ敹鏉ヨ嚜娓叉煋杩涚▼鐨勬棩蹇?
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
