import React, { useState, useEffect, useRef } from 'react'
import axios from 'axios'
import { AlertCircle, FileText, Activity, Settings, Play, Square, Download, LogOut } from 'lucide-react'
import StatsCard from './StatsCard'
import AlertList from './AlertList'
import LogFilesPanel from './LogFilesPanel'
import RulesPanel from './RulesPanel'
import AnalyzePanel from './AnalyzePanel'
import './Dashboard.css'

const API_BASE = '/api'

function Dashboard({ onLogout }) {
  const [activeTab, setActiveTab] = useState('dashboard')
  const [stats, setStats] = useState({
    totalAlerts: 0,
    severityCount: {},
    activeRules: 0,
    watchedFiles: 0
  })
  const [alerts, setAlerts] = useState([])
  const [logFiles, setLogFiles] = useState([])
  const [rules, setRules] = useState([])
  const [isTailing, setIsTailing] = useState(false)
  const wsRef = useRef(null)

  const connectWebSocket = () => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      return
    }
    
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const token = localStorage.getItem('authToken') || sessionStorage.getItem('authToken')
    const wsUrl = token
      ? `${protocol}//${window.location.host}${API_BASE}/tail/ws?token=${encodeURIComponent(token)}`
      : `${protocol}//${window.location.host}${API_BASE}/tail/ws`
    const ws = new WebSocket(wsUrl)
    wsRef.current = ws

    ws.onopen = () => {}

    ws.onmessage = (event) => {
      try {
        const alert = JSON.parse(event.data)
        if (alert.line === undefined || alert.source === undefined) return
        setAlerts(prev => {
          const isNew = !prev.some(a => 
            a.line === alert.line && 
            a.source === alert.source &&
            Math.abs(new Date(a.timestamp) - new Date(alert.timestamp)) < 5000
          )
          
          if (isNew) {
            const updated = [...prev, alert]
            const newAlerts = updated.slice(-1000)
            const sev = alert.severity?.toLowerCase()
            if (sev === 'critical' || sev === 'kritik' || sev === 'high' || sev === 'yüksek') {
              playAlertSound(alert.severity)
            }
            
            return newAlerts
          }
          return prev
        })
        loadStats()
      } catch (err) {}
    }

    ws.onerror = () => {}

    ws.onclose = () => {
      wsRef.current = null
      if (isTailing) {
        setTimeout(() => {
          if (isTailing && !wsRef.current) {
            connectWebSocket()
          }
        }, 2000)
      }
    }
  }

  useEffect(() => {
    const checkTailingStatus = async () => {
      try {
        const res = await axios.get(`${API_BASE}/stats`)
        if (res.data.isTailing) {
          setIsTailing(true)
        }
      } catch (err) {}
    }
    checkTailingStatus()
  }, [])

  useEffect(() => {
    if (isTailing && (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN)) {
      connectWebSocket()
    } else if (!isTailing && wsRef.current) {
      if (wsRef.current.readyState === WebSocket.OPEN || wsRef.current.readyState === WebSocket.CONNECTING) {
        wsRef.current.close()
      }
      wsRef.current = null
    }
  }, [isTailing])

  useEffect(() => {
    loadStats()
    loadLogFiles()
    loadRules()
    loadAlerts()

    const interval = setInterval(() => {
      loadStats()
      loadAlerts()
    }, 3000)

    return () => clearInterval(interval)
  }, [isTailing])

  const loadStats = async () => {
    try {
      const res = await axios.get(`${API_BASE}/stats`)
      setStats(res.data)
      if (res.data.isTailing !== isTailing) {
        setIsTailing(res.data.isTailing)
      }
    } catch (err) {}
  }

  const loadAlerts = async () => {
    try {
      const res = await axios.get(`${API_BASE}/tail/alerts`)
      setAlerts(res.data)
    } catch (err) {}
  }

  const loadLogFiles = async () => {
    try {
      const res = await axios.get(`${API_BASE}/logfiles`)
      setLogFiles(res.data)
    } catch (err) {}
  }

  const loadRules = async () => {
    try {
      const res = await axios.get(`${API_BASE}/rules`)
      setRules(res.data)
    } catch (err) {}
  }

  const startTailing = async () => {
    try {
      const enabledFiles = logFiles.filter(f => f.enabled).map(f => f.path)
      if (enabledFiles.length === 0) {
        alert('Lütfen en az bir log dosyasını aktif hale getirin!')
        return
      }
      const res = await axios.post(`${API_BASE}/tail/start`, { files: enabledFiles })
      setIsTailing(true)
      const started = res.data?.started?.length ?? 0
      const failed = res.data?.failed ?? []
      if (failed.length > 0) {
        alert(`${started}/${enabledFiles.length} dosya izleniyor. ${failed.length} dosya açılamadı (yok veya izin yok).`)
      }
    } catch (err) {
      alert('İzleme başlatılamadı: ' + (err.response?.data?.error || err.message))
    }
  }

  const stopTailing = async () => {
    try {
      const enabledFiles = logFiles.filter(f => f.enabled).map(f => f.path)
      await axios.post(`${API_BASE}/tail/stop`, { files: enabledFiles })
      setIsTailing(false)
    } catch (err) {}
  }

  const playAlertSound = (severity) => {
    try {
      const audioContext = new (window.AudioContext || window.webkitAudioContext)()
      const oscillator = audioContext.createOscillator()
      const gainNode = audioContext.createGain()
      
      oscillator.connect(gainNode)
      gainNode.connect(audioContext.destination)
      const sev = severity?.toLowerCase()
      if (sev === 'critical' || sev === 'kritik') {
        oscillator.frequency.value = 800
        oscillator.type = 'sine'
        gainNode.gain.setValueAtTime(0.3, audioContext.currentTime)
        gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.5)
        oscillator.start(audioContext.currentTime)
        oscillator.stop(audioContext.currentTime + 0.5)
        setTimeout(() => {
          const osc2 = audioContext.createOscillator()
          const gain2 = audioContext.createGain()
          osc2.connect(gain2)
          gain2.connect(audioContext.destination)
          osc2.frequency.value = 800
          osc2.type = 'sine'
          gain2.gain.setValueAtTime(0.3, audioContext.currentTime)
          gain2.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.5)
          osc2.start(audioContext.currentTime)
          osc2.stop(audioContext.currentTime + 0.5)
        }, 300)
      } else if (sev === 'high' || sev === 'yüksek') {
        oscillator.frequency.value = 600
        oscillator.type = 'sine'
        gainNode.gain.setValueAtTime(0.2, audioContext.currentTime)
        gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.3)
        oscillator.start(audioContext.currentTime)
        oscillator.stop(audioContext.currentTime + 0.3)
      }
    } catch (err) {}
  }

  const severityToLabel = (s) => {
    const x = s?.toLowerCase?.() ?? ''
    if (x === 'critical' || x === 'kritik') return 'Kritik'
    if (x === 'high' || x === 'yüksek') return 'Yüksek'
    if (x === 'medium' || x === 'orta') return 'Orta'
    if (x === 'low' || x === 'düşük') return 'Düşük'
    return s || 'Bilinmiyor'
  }

  const exportToCSV = () => {
    const headers = ['Zaman', 'Kaynak', 'LogDosyası', 'Önem', 'Kurallar', 'Özet', 'HamSatır']
    const rows = alerts.map(alert => [
      alert.timestamp,
      alert.source,
      alert.logFile,
      severityToLabel(alert.severity),
      alert.matchedRules?.join('; ') || '',
      alert.summary || alert.line || '',
      alert.line || ''
    ])

    const csvContent = [
      headers.join(','),
      ...rows.map(row => row.map(cell => `"${String(cell).replace(/"/g, '""')}"`).join(','))
    ].join('\n')

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    const url = URL.createObjectURL(blob)
    link.setAttribute('href', url)
    link.setAttribute('download', `log-report-${new Date().toISOString().split('T')[0]}.csv`)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  return (
    <div className="dashboard">
      <header className="dashboard-header">
        <h1>🔍 Log Analiz ve Uyarı Aracı</h1>
        <div className="header-actions">
          {isTailing ? (
            <button className="btn btn-danger" onClick={stopTailing}>
              <Square size={18} /> İzlemeyi Durdur
            </button>
          ) : (
            <button className="btn btn-success" onClick={startTailing}>
              <Play size={18} /> Gerçek Zamanlı İzleme Başlat
            </button>
          )}
          <button className="btn btn-secondary" onClick={exportToCSV}>
            <Download size={18} /> CSV İndir
          </button>
          {onLogout && (
            <button className="btn btn-secondary" onClick={onLogout}>
              <LogOut size={18} /> Çıkış
            </button>
          )}
        </div>
      </header>

      <div className="dashboard-tabs">
        <button
          className={`tab ${activeTab === 'dashboard' ? 'active' : ''}`}
          onClick={() => setActiveTab('dashboard')}
        >
          <Activity size={18} /> Dashboard
        </button>
        <button
          className={`tab ${activeTab === 'alerts' ? 'active' : ''}`}
          onClick={() => setActiveTab('alerts')}
        >
          <AlertCircle size={18} /> Uyarılar ({alerts.length})
        </button>
        <button
          className={`tab ${activeTab === 'analyze' ? 'active' : ''}`}
          onClick={() => setActiveTab('analyze')}
        >
          <FileText size={18} /> Analiz
        </button>
        <button
          className={`tab ${activeTab === 'files' ? 'active' : ''}`}
          onClick={() => setActiveTab('files')}
        >
          <FileText size={18} /> Log Dosyaları
        </button>
        <button
          className={`tab ${activeTab === 'rules' ? 'active' : ''}`}
          onClick={() => setActiveTab('rules')}
        >
          <Settings size={18} /> Kurallar
        </button>
      </div>

      <div className="dashboard-content">
        {activeTab === 'dashboard' && (
          <div className="dashboard-view">
            <div className="stats-grid">
              <StatsCard
                title="Toplam Uyarı"
                value={stats.totalAlerts}
                color="#ef4444"
              />
              <StatsCard
                title="Kritik"
                value={(stats.severityCount?.critical ?? stats.severityCount?.kritik) || 0}
                color="#dc2626"
              />
              <StatsCard
                title="Yüksek"
                value={(stats.severityCount?.high ?? stats.severityCount?.yüksek) || 0}
                color="#f59e0b"
              />
              <StatsCard
                title="Orta"
                value={(stats.severityCount?.medium ?? stats.severityCount?.orta) || 0}
                color="#3b82f6"
              />
              <StatsCard
                title="Aktif Kurallar"
                value={stats.activeRules}
                color="#10b981"
              />
              <StatsCard
                title="İzlenen Dosyalar"
                value={stats.watchedFiles}
                color="#8b5cf6"
              />
            </div>
            {(stats.watchedFiles > 0 || (stats.watchedFilesList && stats.watchedFilesList.length > 0)) && (
              <div className="watched-files-info">
                <strong>İzlenen dosyalar ({stats.watchedFilesList?.length ?? stats.watchedFiles}):</strong>
                {(stats.watchedFilesList && stats.watchedFilesList.length > 0) ? (
                  <ul>
                    {stats.watchedFilesList.map((path, i) => (
                      <li key={i}>{path}</li>
                    ))}
                  </ul>
                ) : (
                  <p className="watched-files-note">Liste yükleniyor…</p>
                )}
                {logFiles.length > 0 && (logFiles.filter(f => f.enabled).length > (stats.watchedFilesList?.length ?? stats.watchedFiles)) && (
                  <p className="watched-files-note">
                    Yapılandırılan {logFiles.filter(f => f.enabled).length} dosyadan {stats.watchedFilesList?.length ?? stats.watchedFiles} tanesi izleniyor; diğerleri bulunamadı veya erişilemedi.
                  </p>
                )}
              </div>
            )}
            <AlertList alerts={alerts.slice(-20)} />
          </div>
        )}

        {activeTab === 'alerts' && (
          <AlertList alerts={alerts} />
        )}

        {activeTab === 'analyze' && (
          <AnalyzePanel logFiles={logFiles} />
        )}

        {activeTab === 'files' && (
          <LogFilesPanel logFiles={logFiles} />
        )}

        {activeTab === 'rules' && (
          <RulesPanel rules={rules} />
        )}
      </div>
    </div>
  )
}

export default Dashboard
