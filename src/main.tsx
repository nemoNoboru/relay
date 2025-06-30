import React from 'react'
import ReactDOM from 'react-dom/client'
import App from '@app/App'
import './index.css'

// Remove Preload scripts loading
postMessage({ payload: 'removeLoading' }, '*')

// Use contextBridge - only in Electron environment
if (typeof window !== 'undefined' && (window as any).ipcRenderer) {
  (window as any).ipcRenderer.on('main-process-message', (_event: any, message: any) => {
    console.log('Message from main process:', message)
  })
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
) 