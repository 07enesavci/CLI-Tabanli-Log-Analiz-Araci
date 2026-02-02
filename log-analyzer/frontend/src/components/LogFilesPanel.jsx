import React from 'react'
import { FileText, CheckCircle, XCircle } from 'lucide-react'
import './LogFilesPanel.css'

function LogFilesPanel({ logFiles }) {
  return (
    <div className="log-files-panel">
      <div className="log-files-panel-header">
        <h2>Log Dosyaları</h2>
      </div>

      <div className="log-files-grid">
        {logFiles.map((file, index) => (
          <div key={index} className="log-file-card">
            <div className="log-file-header">
              <FileText size={24} />
              <div className="log-file-status">
                {file.enabled ? (
                  <CheckCircle size={20} color="#10b981" />
                ) : (
                  <XCircle size={20} color="#ef4444" />
                )}
              </div>
            </div>
            <div className="log-file-path">{file.path}</div>
            <div className="log-file-type">Tip: {file.type}</div>
            <div className="log-file-status-text">
              Durum: {file.enabled ? 'Aktif' : 'Pasif'}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default LogFilesPanel
