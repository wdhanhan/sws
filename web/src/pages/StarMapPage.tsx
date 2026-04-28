import { useState, useEffect } from 'react'
import api from '../api/client'
import { useGameStore } from '../stores/gameStore'

export default function StarMapPage() {
  const { charId } = useGameStore()
  const setSystem = useGameStore(s => s.setSystem)
  const [system, setSystemData] = useState<any>(null)
  const [currentSystemId, setCurrentSystemId] = useState(0)

  const loadSystem = async (id: number) => {
    try {
      const res: any = await api.get(`/starmap/systems/${id}`)
      setSystemData(res.data)
      setSystem(id, res.data.system.name)
      setCurrentSystemId(id)
    } catch {}
  }

  useEffect(() => {
    const fetchChar = async () => {
      try {
        const res: any = await api.get(`/characters/${charId}`)
        loadSystem(res.data.current_system_id)
      } catch {}
    }
    if (charId) fetchChar()
  }, [charId])

  if (!system) return <div className="center dim mt">加载星系数据中...</div>

  const s = system.system
  const secColor = s.security_level >= 0.5 ? 'green' : s.security_level > 0 ? 'gold' : 'red'

  return (
    <div>
      <div className="title">
        ═══ {s.name} ═══{' '}
        <span className={secColor}>[{system.zone === 'high' ? '高安' : system.zone === 'low' ? '低安' : system.zone === 'null' ? '零安' : '深渊'}]</span>
      </div>

      <div className="dim">恒星: {system.star_name} | 旋臂: {system.arm_name}</div>

      <div className="section">
        <div className="section-title">── 行星 ({system.planets?.length || 0}) ──</div>
        {system.planets?.map((p: any) => (
          <div key={p.id} className="item-row">
            <span>{p.name} {p.planet_type}</span>
            <span className="dim">{p.orbit_au}AU {p.has_station && '🏗 空间站'}</span>
          </div>
        ))}
      </div>

      <div className="section">
        <div className="section-title">── 矿带 ({system.belts?.length || 0}) ──</div>
        {system.belts?.map((b: any) => (
          <div key={b.id} className="item-row">
            <span>⛏ {b.name} ({b.belt_type})</span>
            <span className="dim">剩余{b.remaining_pct}%</span>
          </div>
        ))}
      </div>

      <div className="section">
        <div className="section-title">── 星门 → ──</div>
        {system.gates_to?.map((g: any) => (
          <div key={g.id} className="item-row" onClick={() => loadSystem(g.id)}>
            <span className="blue">&gt; {g.name}</span>
            <span className={g.security_level >= 0.5 ? 'green' : g.security_level > 0 ? 'gold' : 'red'}>
              [{g.security_level.toFixed(1)}]
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
