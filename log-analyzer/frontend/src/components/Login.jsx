import React, { useState } from 'react'
import axios from 'axios'
import './Login.css'

const API_BASE = '/api'

function Login({ onSuccess }) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const res = await axios.post(`${API_BASE}/login`, { username, password })
      if (res.data?.success && res.data?.token) {
        localStorage.setItem('authToken', res.data.token)
        sessionStorage.setItem('authToken', res.data.token)
        axios.defaults.headers.common['Authorization'] = `Bearer ${res.data.token}`
        onSuccess()
      } else {
        setError(res.data?.error || 'Giriş başarısız')
      }
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Bağlantı hatası')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-page">
      <div className="login-box">
        <h1>Log Analiz Sistemi</h1>
        <p className="login-subtitle">Giriş yapın</p>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="username">Kullanıcı adı</label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Kullanıcı adı"
              autoComplete="username"
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="password">Şifre</label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Şifre"
              autoComplete="current-password"
              required
            />
          </div>
          {error && <p className="login-error">{error}</p>}
          <button type="submit" className="login-btn" disabled={loading}>
            {loading ? 'Giriş yapılıyor...' : 'Giriş'}
          </button>
        </form>
      </div>
    </div>
  )
}

export default Login
