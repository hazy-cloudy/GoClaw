const { app, BrowserWindow, ipcMain, Menu, screen, globalShortcut } = require('electron');
const path = require('path');

let petWindow = null;
let dashboardWindow = null;
let onboardingWindow = null;
let onboardingMode = false;

const DASHBOARD_URL = process.env.GOCLAW_DASHBOARD_URL || 'http://127.0.0.1:3000';
const LAUNCHER_TOKEN = process.env.GOCLAW_LAUNCHER_TOKEN || '';

function buildDashboardURL(forceOnboarding = false) {
  try {
    const target = new URL(DASHBOARD_URL);
    if (LAUNCHER_TOKEN && !target.searchParams.get('token')) {
      target.searchParams.set('token', LAUNCHER_TOKEN);
    }
    if (forceOnboarding) {
      target.searchParams.set('onboarding', '1');
    }
    return target.toString();
  } catch {
    return DASHBOARD_URL;
  }
}

function bindWindowTitleSync(win, fallbackTitle) {
  win.setTitle(fallbackTitle);
  win.webContents.on('page-title-updated', (event, title) => {
    event.preventDefault();
    const nextTitle = title && title.trim() ? title : fallbackTitle;
    win.setTitle(nextTitle);
  });
}

function hideMainWindowsForOnboarding() {
  onboardingMode = true;

  if (petWindow && !petWindow.isDestroyed()) {
    petWindow.hide();
  }

  if (dashboardWindow && !dashboardWindow.isDestroyed()) {
    dashboardWindow.hide();
  }
}

function restoreMainWindowsAfterOnboarding() {
  onboardingMode = false;

  if (petWindow && !petWindow.isDestroyed()) {
    petWindow.show();
  }
}

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
      sandbox: false,
      preload: path.join(__dirname, 'preload.js')
    }
  });

  petWindow.loadURL('http://localhost:5173');
  
  petWindow.once('ready-to-show', () => {
    if (!onboardingMode) {
      petWindow.show();
    }
  });

  petWindow.on('closed', () => {
    petWindow = null;
    app.quit();
  });

  bindWindowTitleSync(petWindow, 'GoClaw 桌宠');

  return petWindow;
}

function createDashboardWindow() {
  if (dashboardWindow) {
    dashboardWindow.focus();
    return;
  }

  const { width, height } = screen.getPrimaryDisplay().workAreaSize;
  const dashboardWidth = Math.max(1024, Math.round(width * 0.78));
  const dashboardHeight = Math.max(720, Math.round(height * 0.82));

  dashboardWindow = new BrowserWindow({
    width: dashboardWidth,
    height: dashboardHeight,
    minWidth: 960,
    minHeight: 640,
    autoHideMenuBar: true,
    menuBarVisible: false,
    frame: false,
    transparent: true,
    backgroundColor: '#00000000',
    resizable: true,
    minimizable: true,
    maximizable: true,
    closable: true,
    movable: true,
    show: false,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      sandbox: false,
      preload: path.join(__dirname, 'preload.js')
    }
  });

  dashboardWindow.loadURL(buildDashboardURL(false));
  bindWindowTitleSync(dashboardWindow, 'GoClaw 控制台');

  dashboardWindow.setIgnoreMouseEvents(false);

  dashboardWindow.once('ready-to-show', () => {
    dashboardWindow.setIgnoreMouseEvents(false);
    dashboardWindow.show();
  });

  dashboardWindow.on('focus', () => {
    if (!dashboardWindow || dashboardWindow.isDestroyed()) {
      return;
    }
    dashboardWindow.setIgnoreMouseEvents(false);
  });

  dashboardWindow.on('closed', () => {
    dashboardWindow = null;
  });
}

function createOnboardingWindow() {
  if (onboardingWindow) {
    hideMainWindowsForOnboarding();
    onboardingWindow.loadURL(buildDashboardURL(true));
    onboardingWindow.setIgnoreMouseEvents(false);
    onboardingWindow.show();
    onboardingWindow.focus();
    return;
  }

  const { width, height } = screen.getPrimaryDisplay().workAreaSize;
  const onboardingWidth = Math.max(1120, Math.round(width * 0.84));
  const onboardingHeight = Math.max(760, Math.round(height * 0.86));

  onboardingWindow = new BrowserWindow({
    width: onboardingWidth,
    height: onboardingHeight,
    minWidth: 1024,
    minHeight: 700,
    autoHideMenuBar: true,
    menuBarVisible: false,
    frame: false,
    transparent: true,
    backgroundColor: '#00000000',
    resizable: true,
    minimizable: true,
    maximizable: true,
    closable: true,
    movable: true,
    show: false,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      sandbox: false,
      preload: path.join(__dirname, 'preload.js')
    }
  });

  hideMainWindowsForOnboarding();

  onboardingWindow.loadURL(buildDashboardURL(true));
  bindWindowTitleSync(onboardingWindow, 'GoClaw 初始化向导');

  onboardingWindow.setIgnoreMouseEvents(false);

  onboardingWindow.once('ready-to-show', () => {
    onboardingWindow.setIgnoreMouseEvents(false);
    onboardingWindow.show();
  });

  onboardingWindow.on('closed', () => {
    onboardingWindow = null;
    restoreMainWindowsAfterOnboarding();
  });
}


ipcMain.on('open-dashboard', () => {
  if (onboardingMode && onboardingWindow && !onboardingWindow.isDestroyed()) {
    onboardingWindow.show();
    onboardingWindow.focus();
    return;
  }
  createDashboardWindow();
});

ipcMain.on('open-onboarding', () => {
  createOnboardingWindow();
});

ipcMain.on('set-onboarding-mode', (event, enabled) => {
  const senderWindow = BrowserWindow.fromWebContents(event.sender);
  if (enabled) {
    hideMainWindowsForOnboarding();
    if (senderWindow && !senderWindow.isDestroyed()) {
      senderWindow.show();
      senderWindow.focus();
    }
    return;
  }

  restoreMainWindowsAfterOnboarding();
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

ipcMain.on('set-click-through', (event, enabled) => {
  if (petWindow && !petWindow.isDestroyed()) {
    petWindow.setOpacity(enabled ? 0.2 : 1);
    petWindow.setIgnoreMouseEvents(enabled, { forward: true });
  }
});

app.whenReady().then(() => {
  Menu.setApplicationMenu(null);
  globalShortcut.register('F12', () => {
    const focused = BrowserWindow.getFocusedWindow();
    if (focused && !focused.isDestroyed()) {
      focused.webContents.toggleDevTools();
    }
  });
  globalShortcut.register('CommandOrControl+Shift+I', () => {
    const focused = BrowserWindow.getFocusedWindow();
    if (focused && !focused.isDestroyed()) {
      focused.webContents.toggleDevTools();
    }
  });
  createPetWindow();
});

app.on('will-quit', () => {
  globalShortcut.unregisterAll();
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit();
});
