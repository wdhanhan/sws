import { useState, useEffect, useRef, useCallback } from 'react'
import api from '../api/client'
import HealthBar from '../components/HealthBar'

type View = 'sites' | 'fighting' | 'result'

const siteLabels: Record<string, string> = {
  small_anomaly: '小型异常', medium_anomaly: '中型异常', large_anomaly: '大型异常',
  signal: '战斗信号', expedition: '远征入口',
}

export default function CombatPage() {
  const [view, setView] = useState<View>('sites')
  const [sites, setSites] = useState<any[]>([])
  const [result, setResult] = useState<any>(null)
  const [autoMode, setAutoMode] = useState(false)
  const autoRef = useRef(false)
  const timerRef = useRef<any>(null)
  const logRef = useRef<HTMLDivElement>(null)

  useEffect(() => { autoRef.current = autoMode }, [autoMode])
  useEffect(() => { return () => { if (timerRef.current) clearTimeout(timerRef.current) } }, [])
  useEffect(() => { logRef.current?.scrollIntoView({ behavior: 'smooth' }) }, [result])

  const loadSites = useCallback(async () => {
    try {
      const res: any = await api.get('/sites?show_all=true')
      setSites(res.data.sites || [])
    } catch {}
  }, [])

  // 进入页面时检查是否有进行中的战斗
  useEffect(() => {
    const checkSession = async () => {
      try {
        const res: any = await api.post('/sites/tick')
        if (res.data && res.data.combat && res.data.combat.status === 'active') {
          setResult(res.data)
          setView('fighting')
          return
        }
        if (res.data && (res.data.completed || res.data.failed)) {
          setResult(res.data)
          setView('result')
          return
        }
      } catch {}
      loadSites()
    }
    checkSession()
  }, [loadSites])

  const scan = async () => {
    const res: any = await api.post('/sites/scan')
    setSites(res.data.sites || [])
  }

  const enter = async (siteID: number) => {
    try {
      const res: any = await api.post('/sites/enter', { site_id: siteID })
      setResult(res.data)
      setView('fighting')
    } catch (e: any) { alert(e?.message || '无法进入') }
  }

  const tick = async (): Promise<boolean> => {
    try {
      const res: any = await api.post('/sites/tick')
      setResult(res.data)
      if (res.data.completed || res.data.failed) { setView('result'); return false }
      return true
    } catch { return false }
  }

  const startAuto = () => {
    autoRef.current = true
    setAutoMode(true)
    const loop = async () => {
      if (!autoRef.current) return
      const ok = await tick()
      if (ok && autoRef.current) {
        timerRef.current = setTimeout(loop, 1000)
      } else {
        autoRef.current = false
        setAutoMode(false)
      }
    }
    loop()
  }

  const stopAuto = () => {
    autoRef.current = false
    setAutoMode(false)
    if (timerRef.current) clearTimeout(timerRef.current)
  }

  const leave = async () => {
    stopAuto()
    try { await api.post('/sites/leave') } catch {}
    setResult(null)
    setView('sites')
    loadSites()
  }

  // ========== 地点列表 ==========
  if (view === 'sites') {
    return (
      <div>
        <div className="title">═══ 战斗地点 ═══</div>
        <div className="center mb">
          <button className="btn" onClick={scan}>[ 扫描本星系 ]</button>
          <button className="btn dim" onClick={loadSites}>[ 刷新 ]</button>
          <button className="btn dim" onClick={async () => {
            try { await api.post('/sites/leave') } catch {}
            setResult(null)
          }}>[ 清除残留 ]</button>
        </div>
        {sites.length === 0 && <div className="dim center">暂无地点，点击扫描...</div>}
        {sites.map((s: any) => (
          <div key={s.id} className="item-row" onClick={() => enter(s.id)}>
            <div>
              <span className={s.difficulty <= 2 ? 'green' : s.difficulty <= 5 ? 'gold' : 'red'}>
                [{siteLabels[s.site_type] || s.site_type}]
              </span>{' '}
              <span className="blue">{s.name}</span>
              <div className="dim" style={{ fontSize: 11 }}>难度 T{s.difficulty}</div>
            </div>
            <span className={s.difficulty <= 2 ? 'green' : s.difficulty <= 5 ? 'gold' : 'red'}>T{s.difficulty}</span>
          </div>
        ))}
      </div>
    )
  }

  // ========== 战斗进行中 ==========
  if (view === 'fighting' && result) {
    const r = result
    const combat = r.combat || {}
    const logs = combat.logs || []
    const parts = combat.participants || []
    const player = parts.find((p: any) => p.type === 'player')
    const enemies = parts.filter((p: any) => p.type === 'npc')

    return (
      <div>
        <div className="title">
          ═══ {r.site_name || '战斗地点'} ═══ 第{r.wave_number}/{r.total_waves}波
          {r.is_boss && <span className="red"> [BOSS {r.boss_name}]</span>}
          <span className="dim"> Tick {combat.tick || 0}</span>
        </div>

        {r.wave_text && <div className="dim" style={{ padding: '4px 0', fontSize: 12 }}>{r.wave_text}</div>}

        {/* 玩家状态 */}
        {player && (
          <div className="section" style={{ padding: '6px 0' }}>
            <div className="blue">你 [{player.name}] 距离:{(player.distance / 1000).toFixed(0)}km</div>
            <div><HealthBar current={player.shield_current} max={player.shield_max} label="盾" /></div>
            <div><HealthBar current={player.armor_current} max={player.armor_max} label="甲" /></div>
            <div><HealthBar current={player.structure_current} max={player.structure_max} label="构" /></div>
          </div>
        )}

        {/* 敌方状态(每个敌人都显示三层HP) */}
        {enemies.map((e: any, i: number) => (
          <div key={i} style={{ padding: '3px 0', borderBottom: '1px solid var(--border)' }}>
            <span className={e.is_destroyed ? 'dim' : 'red'}>
              {e.is_destroyed ? '✗' : '▸'} {e.name} {!e.is_destroyed && `${(e.distance / 1000).toFixed(0)}km`}
            </span>
            {!e.is_destroyed && (
              <div style={{ fontSize: 12, paddingLeft: 12 }}>
                <HealthBar current={e.shield_current} max={e.shield_max} label="盾" width={8} />
                <HealthBar current={e.armor_current} max={e.armor_max} label="甲" width={8} />
                <HealthBar current={e.structure_current} max={e.structure_max} label="构" width={8} />
              </div>
            )}
          </div>
        ))}

        {/* 战报 */}
        <div style={{ maxHeight: 160, overflowY: 'auto', fontSize: 11, margin: '6px 0', padding: 4, background: 'var(--bg)' }}>
          {logs.slice(-15).map((l: string, i: number) => (
            <div key={i}>
              <span className={l.includes('击毁') || l.includes('通关') ? 'epic' : l.includes('命中') ? 'damage' : l.includes('Tick') ? 'dim' : 'info'}>{l}</span>
            </div>
          ))}
          <div ref={logRef} />
        </div>

        {/* 控制 */}
        <div className="center">
          {!autoMode ? (
            <>
              <button className="btn" onClick={() => tick()}>[ 下一Tick ]</button>
              <button className="btn btn-success" onClick={startAuto}>[ ▶ 自动(1秒/Tick) ]</button>
            </>
          ) : (
            <button className="btn btn-danger" onClick={stopAuto}>[ ■ 停止自动 ]</button>
          )}
          <button className="btn btn-danger" onClick={leave}>[ 撤退 ]</button>
        </div>
        {autoMode && <div className="dim center" style={{ fontSize: 11 }}>自动战斗中，每秒推进1Tick...</div>}
      </div>
    )
  }

  // ========== 结果页 ==========
  return (
    <div>
      <div className="title">═══ {result?.completed ? '通关!' : '战斗结束'} ═══</div>
      {result?.combat?.logs?.slice(-12).map((l: string, i: number) => (
        <div key={i} className="log-line">
          <span className={l.includes('通关') || l.includes('战利品') ? 'epic' : l.includes('击毁') ? 'red' : 'info'}>{l}</span>
        </div>
      ))}
      {result?.rewards?.map((r: string, i: number) => (
        <div key={i} className="gold">★ {r}</div>
      ))}
      <div className="center mt">
        <button className="btn" onClick={() => { setResult(null); setView('sites'); loadSites() }}>[ 返回地点列表 ]</button>
      </div>
    </div>
  )
}
