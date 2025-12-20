
import React from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import { registerSW } from 'virtual:pwa-register'

registerSW({
  immediate: true,
  onOfflineReady() {
    console.log('AvidLearner is ready to work offline.')
  },
  onNeedRefresh() {
    console.log('New content available, please refresh.')
  },
  onRegistered(registration) {
    console.log('Service Worker registered:', registration)
  },
  onRegisterError(error) {
    console.error('Service Worker registration failed:', error)
  }
})

createRoot(document.getElementById('root')).render(<App />)
