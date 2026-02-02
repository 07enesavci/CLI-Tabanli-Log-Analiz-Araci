import React from 'react'
import './StatsCard.css'

function StatsCard({ title, value, color }) {
  return (
    <div className="stats-card" style={{ borderTopColor: color }}>
      <div className="stats-card-title">{title}</div>
      <div className="stats-card-value" style={{ color }}>
        {value.toLocaleString()}
      </div>
    </div>
  )
}

export default StatsCard
