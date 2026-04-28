interface Props {
  current: number
  max: number
  label?: string
  width?: number
}

export default function HealthBar({ current, max, label, width = 10 }: Props) {
  const pct = max > 0 ? Math.round((current / max) * 100) : 0
  const filled = Math.round((pct / 100) * width)
  const empty = width - filled
  const colorClass = pct > 50 ? 'full' : pct > 25 ? 'full' : 'low'

  return (
    <span className="hp-bar">
      {label && <span>{label} </span>}
      <span className="bar">
        <span className={colorClass}>{'█'.repeat(filled)}</span>
        <span className="empty">{'░'.repeat(empty)}</span>
      </span>
      {' '}{pct}%
    </span>
  )
}
