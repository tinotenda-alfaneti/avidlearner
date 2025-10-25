
import React from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import { registerSW } from 'virtual:pwa-register'

registerSW({
  immediate: true,
  onOfflineReady() {
    console.log('AvidLearner is ready to work offline.')
  }
})

createRoot(document.getElementById('root')).render(<App />)
