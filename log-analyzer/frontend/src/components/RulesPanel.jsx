import React from 'react'
import { Shield, CheckCircle, XCircle } from 'lucide-react'
import './RulesPanel.css'

function RulesPanel({ rules }) {
  const getSeverityColor = (severity) => {
    const s = severity?.toLowerCase()
    if (s === 'critical' || s === 'kritik') return '#dc2626'
    if (s === 'high' || s === 'yüksek') return '#f59e0b'
    if (s === 'medium' || s === 'orta') return '#3b82f6'
    if (s === 'low' || s === 'düşük') return '#10b981'
    return '#6b7280'
  }

  const severityToLabel = (severity) => {
    const s = severity?.toLowerCase?.() ?? ''
    if (s === 'critical' || s === 'kritik') return 'Kritik'
    if (s === 'high' || s === 'yüksek') return 'Yüksek'
    if (s === 'medium' || s === 'orta') return 'Orta'
    if (s === 'low' || s === 'düşük') return 'Düşük'
    return severity || 'Bilinmiyor'
  }

  return (
    <div className="rules-panel">
      <div className="rules-panel-header">
        <h2>Kurallar</h2>
      </div>

      <div className="rules-grid">
        {rules.map((rule, index) => (
          <div key={index} className="rule-card">
            <div className="rule-header">
              <Shield size={24} />
              <div className="rule-status">
                {rule.enabled ? (
                  <CheckCircle size={20} color="#10b981" />
                ) : (
                  <XCircle size={20} color="#ef4444" />
                )}
              </div>
            </div>
            <div className="rule-name">{rule.name}</div>
            <div className="rule-severity" style={{ color: getSeverityColor(rule.severity) }}>
              Önem: {severityToLabel(rule.severity)}
            </div>
            <div className="rule-description">{rule.description}</div>
            <div className="rule-pattern">
              <strong>Desen:</strong> <code>{rule.pattern}</code>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default RulesPanel
