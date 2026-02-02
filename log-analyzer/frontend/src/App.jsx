import React, { useState } from 'react'
import axios from 'axios'
import Dashboard from './components/Dashboard'
import Login from './components/Login'
import './App.css'

const AUTH_TOKEN_KEY = 'authToken'

function getStoredToken() {
  try {
    return localStorage.getItem(AUTH_TOKEN_KEY) || sessionStorage.getItem(AUTH_TOKEN_KEY)
  } catch {
    return null
  }
}

function clearStoredToken() {
  localStorage.removeItem(AUTH_TOKEN_KEY)
  sessionStorage.removeItem(AUTH_TOKEN_KEY)
}

function App() {
  const [authenticated, setAuthenticated] = useState(() => {
    const token = getStoredToken()
    if (token) {
      axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
      return true
    }
    return false
  })

  const handleLoginSuccess = () => {
    setAuthenticated(true)
  }

  const handleLogout = () => {
    clearStoredToken()
    delete axios.defaults.headers.common['Authorization']
    setAuthenticated(false)
  }

  if (!authenticated) {
    return (
      <div className="App">
        <Login onSuccess={handleLoginSuccess} />
      </div>
    )
  }

  return (
    <div className="App">
      <Dashboard onLogout={handleLogout} />
    </div>
  )
}

export default App
