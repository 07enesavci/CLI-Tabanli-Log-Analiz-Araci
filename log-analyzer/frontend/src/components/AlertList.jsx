import React from 'react'
import { AlertCircle } from 'lucide-react'
import './AlertList.css'

function severityToLabel(severity) {
  const s = severity?.toLowerCase?.() ?? ''
  if (s === 'critical' || s === 'kritik') return 'Kritik'
  if (s === 'high' || s === 'yüksek') return 'Yüksek'
  if (s === 'medium' || s === 'orta') return 'Orta'
  if (s === 'low' || s === 'düşük') return 'Düşük'
  return severity || 'Bilinmiyor'
}

function AlertList({ alerts }) {
  const getSeverityColor = (severity) => {
    const s = severity?.toLowerCase()
    if (s === 'critical' || s === 'kritik') return '#dc2626'
    if (s === 'high' || s === 'yüksek') return '#f59e0b'
    if (s === 'medium' || s === 'orta') return '#3b82f6'
    if (s === 'low' || s === 'düşük') return '#10b981'
    return '#6b7280'
  }

  const formatTime = (timestamp) => {
    if (!timestamp) return 'N/A'
    const date = new Date(timestamp)
    return date.toLocaleString('tr-TR')
  }

  if (alerts.length === 0) {
    return (
      <div className="alert-list-empty">
        <AlertCircle size={48} />
        <p>Henüz uyarı bulunmuyor</p>
      </div>
    )
  }

  return (
    <div className="alert-list">
      <div className="alert-list-header">
        <h2>Uyarılar ({alerts.length})</h2>
      </div>
      <div className="alert-list-items">
        {alerts.slice().reverse().map((alert, index) => (
          <div key={index} className="alert-item">
            <div className="alert-item-header">
              <div className="alert-severity" style={{ backgroundColor: getSeverityColor(alert.severity) }}>
                {severityToLabel(alert.severity)}
              </div>
              <div className="alert-time">{formatTime(alert.timestamp)}</div>
            </div>
            <div className="alert-rules">
              <strong>Kurallar:</strong> {alert.matchedRules?.join(', ') || 'N/A'}
            </div>
            <div className="alert-source">
              <strong>Kaynak:</strong> {alert.source || alert.logFile || 'N/A'}
            </div>
            <div className="alert-summary">
              {alert.summary || alert.line || 'N/A'}
            </div>
            {alert.summary && alert.line && alert.line !== alert.summary && (
              <details className="alert-raw">
                <summary>Ham log satırı</summary>
                <div className="alert-line">{alert.line}</div>
              </details>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}

export default AlertList
