const { app, BrowserWindow, ipcMain, screen } = require('electron');
const path = require('path');

let petWindow = null;
let settingsWindow = null;

function createPetWindow() {
  // 获取屏幕右下角位置
  const { width, height } = screen.getPrimaryDisplay().workAreaSize;
  
  // 桌宠窗口大小
  const petWidth = 280;
  const petHeight = 380;
  
  petWindow = new BrowserWindow({
    width: petWidth,
    height: petHeight,
    x: width - petWidth - 20,
    y: height - petHeight - 60,
    frame: false,           // 无边框
    transparent: true,      // 透明背景
    alwaysOnTop: true,     // 置顶
    resizable: false,       // 不可调整大小
    skipTaskbar: true,     // 不显示在任务栏
    show: false,           // 加载完成后显示
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js')
    }
  });

  petWindow.loadURL('http://localhost:5173');
  
  petWindow.once('ready-to-show', () => {
    petWindow.show();
  });

  petWindow.on('closed', () => {
    petWindow = null;
    app.quit();
  });
}

function createSettingsWindow() {
  if (settingsWindow) {
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
    transparent: true,
    alwaysOnTop: false,
    autoHideMenuBar: true,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js')
    }
  });

  settingsWindow.loadURL('http://localhost:5173/settings.html');

  settingsWindow.once('ready-to-show', () => {
    settingsWindow.show();
  });

  settingsWindow.on('closed', () => {
    settingsWindow = null;
  });
}

ipcMain.on('open-settings', () => {
  createSettingsWindow();
});

ipcMain.on('minimize-settings', () => {
  if (settingsWindow) settingsWindow.minimize();
});

ipcMain.on('maximize-settings', () => {
  if (settingsWindow) {
    if (settingsWindow.isMaximized()) {
      settingsWindow.unmaximize();
    } else {
      settingsWindow.maximize();
    }
  }
});

ipcMain.on('close-settings', () => {
  if (settingsWindow) settingsWindow.close();
});

ipcMain.on('settings-changed', (event, settings) => {
  if (petWindow && !petWindow.isDestroyed()) {
    petWindow.webContents.send('settings-updated', settings);
  }
});

ipcMain.on('chat-history', (event, history) => {
  if (settingsWindow && !settingsWindow.isDestroyed()) {
    settingsWindow.webContents.send('chat-history-updated', history);
  }
});

ipcMain.on('show-bubble', (event, data) => {
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