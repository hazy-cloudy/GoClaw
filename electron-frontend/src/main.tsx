import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import Settings from './settings'
import './index.css'

const hash = window.location.hash
const isSettings = hash === '#settings'

console.log('[App] hash:', hash, 'isSettings:', isSettings)

const root = ReactDOM.createRoot(document.getElementById('root')!)

if (isSettings) {
  root.render(<Settings />)
} else {
  root.render(<App />)
}