import React, { useState } from 'react'
import axios from 'axios'
import { Search, Loader } from 'lucide-react'
import AlertList from './AlertList'
import './AnalyzePanel.css'

const API_BASE = '/api'

function AnalyzePanel({ logFiles }) {
  const [selectedFiles, setSelectedFiles] = useState([])
  const [analyzing, setAnalyzing] = useState(false)
  const [results, setResults] = useState([])

  const toggleFile = (filePath) => {
    setSelectedFiles(prev => {
      if (prev.includes(filePath)) {
        return prev.filter(f => f !== filePath)
      }
      return [...prev, filePath]
    })
  }

  const analyze = async () => {
    if (selectedFiles.length === 0) {
      alert('Lütfen en az bir dosya seçin!')
      return
    }

    setAnalyzing(true)
    try {
      const res = await axios.post(`${API_BASE}/analyze`, {
        files: selectedFiles
      })
      setResults(res.data.entries || [])
    } catch (err) {
      alert('Analiz hatası: ' + err.message)
    } finally {
      setAnalyzing(false)
    }
  }

  return (
    <div className="analyze-panel">
      <h2>Log Dosyası Analizi</h2>
      
      <div className="analyze-controls">
        <div className="file-selection">
          <h3>Analiz Edilecek Dosyaları Seçin</h3>
          <div className="file-list">
            {logFiles.filter(f => f.enabled).map((file, index) => (
              <label key={index} className="file-checkbox">
                <input
                  type="checkbox"
                  checked={selectedFiles.includes(file.path)}
                  onChange={() => toggleFile(file.path)}
                />
                <span>{file.path}</span>
                <span className="file-type-badge">{file.type}</span>
              </label>
            ))}
          </div>
        </div>

        <button
          className="btn-analyze"
          onClick={analyze}
          disabled={analyzing || selectedFiles.length === 0}
        >
          {analyzing ? (
            <>
              <Loader size={18} className="spinner" /> Analiz Ediliyor...
            </>
          ) : (
            <>
              <Search size={18} /> Analiz Et
            </>
          )}
        </button>
      </div>

      {results.length > 0 && (
        <div className="analyze-results">
          <div className="results-header">
            <h3>Analiz Sonuçları ({results.length} uyarı bulundu)</h3>
          </div>
          <AlertList alerts={results} />
        </div>
      )}
    </div>
  )
}

export default AnalyzePanel
