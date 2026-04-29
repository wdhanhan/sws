import { useState, useEffect, useRef, useCallback } from 'react'
import api from '../api/client'
import HealthBar from '../components/HealthBar'

type View = 'sites' | 'fighting' | 'result'

const siteLabels: Record<string, string> = {
  small_anomaly: '小型异常', medium_anomaly: '中型异常', large_anomaly: '大型异常',
  signal: '战斗信号', expedition: '远征入口',
}

function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null)
  const listenersRef = useRef<((msg: any) => void)[]>([])

  const connect = useCallback(() => {
    if (wsRef.current && wsRef.current.readyState <= 1) return wsRef.current

    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${proto}//${location.host}/ws/sites`)
    wsRef.current = ws

    ws.onopen = () => {
      const token = localStorage.getItem('token') || ''
      const charId = parseInt(localStorage.getItem('charId') || '0', 10)
      ws.send(JSON.stringify({ type: 'auth', data: { token, char_id: charId } }))
    }

    ws.onmessage = (ev) => {
      try {
        const msg = JSON.parse(ev.data)
        listenersRef.current.forEach(fn => fn(msg))
      } catch {}
    }

    ws.onclose = () => { wsRef.current = null }
    return ws
  }, [])

  const send = useCallback((type: string, data?: any) => {
    const ws = wsRef.current
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type, data }))
    }
  }, [])

  const close = useCallback(() => {
    wsRef.current?.close()
    wsRef.current = null
  }, [])

  const onMessage = useCallback((fn: (msg: any) => void) => {
    listenersRef.current.push(fn)
    return () => { listenersRef.current = listenersRef.current.filter(f => f !== fn) }
  }, [])

  return { connect, send, close, onMessage, wsRef }
}

export default function CombatPage() {
  const [view, setView] = useState<View>('sites')
  const [sites, setSites] = useState<any[]>([])
  const [result, setResult] = useState<any>(null)
  const [autoMode, setAutoMode] = useState(false)
  const autoRef = useRef(false)
  const logRef = useRef<HTMLDivElement>(null)
  const { connect, send, close, onMessage } = useWebSocket()
  const [wsReady, setWsReady] = useState(false)

  useEffect(() => { autoRef.current = autoMode }, [autoMode])
  useEffect(() => { if (logRef.current) logRef.current.scrollTop = logRef.current.scrollHeight }, [result])

  const loadSites = useCallback(async () => {
    try {
      const res: any = await api.get('/sites?show_all=true')
      setSites(res.data.sites || [])
    } catch {}
  }, [])

  // WS message handler
  useEffect(() => {
    const unsub = onMessage((msg: any) => {
      switch (msg.type) {
        case 'authed':
          setWsReady(true)
          break
        case 'entered':
        case 'tick':
          if (msg.data) {
            setResult(msg.data)
            if (msg.data.completed || msg.data.failed) {
              setView('result')
              setAutoMode(false)
              autoRef.current = false
            } else if (msg.data.combat) {
              setView('fighting')
            }
          }
          break
        case 'auto_stopped':
          setAutoMode(false)
          autoRef.current = false
          break
        case 'left':
          setResult(null)
          setView('sites')
          loadSites()
          break
        case 'error':
          break
      }
    })
    return unsub
  }, [onMessage, loadSites])

  // On mount: load sites
  useEffect(() => { loadSites() }, [loadSites])

  const connectAndEnter = async (siteID: number) => {
    connect()
    // Wait for auth then send enter
    const unsub = onMessage((msg: any) => {
      if (msg.type === 'authed') {
        send('enter', { site_id: siteID })
        unsub()
      }
    })
  }

  const enterSite = async (siteID: number) => {
    if (wsReady) {
      send('enter', { site_id: siteID })
    } else {
      connectAndEnter(siteID)
    }
  }

  const tick = () => send('tick')
  const startAuto = () => {
    autoRef.current = true
    setAutoMode(true)
    send('auto_start')
  }
  const stopAuto = () => {
    autoRef.current = false
    setAutoMode(false)
    send('auto_stop')
  }
  const leave = () => {
    stopAuto()
    send('leave')
  }

  // Cleanup WS on unmount
  useEffect(() => { return () => { close() } }, [close])

  const scan = async () => {
    const res: any = await api.post('/sites/scan')
    setSites(res.data.sites || [])
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
          <div key={s.id} className="item-row" onClick={() => enterSite(s.id)}>
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
    const players = parts.filter((p: any) => p.type === 'player')
    const enemies = parts.filter((p: any) => p.type === 'npc')

    return (
      <div>
        <div className="title">
          ═══ {r.site_name || '战斗地点'} ═══ 第{r.wave_number}/{r.total_waves}波
          {r.is_boss && <span className="red"> [BOSS {r.boss_name}]</span>}
          <span className="dim"> Tick {combat.tick || 0}</span>
          {players.length > 1 && <span className="blue" style={{fontSize:10}}> 舰队x{players.length}</span>}
        </div>

        {r.wave_text && <div className="dim" style={{ padding: '2px 0', fontSize: 11 }}>{r.wave_text}</div>}

        {/* Friendly fleet */}
        <div style={{ borderBottom: '1px solid var(--border)', paddingBottom: 4, marginBottom: 4 }}>
          <div className="dim" style={{ fontSize: 10 }}>▸ 友方 ({players.filter((p:any)=>!p.is_destroyed).length}/{players.length})</div>
          {players.map((p: any, pi: number) => {
            const hpPct = p.shield_max + p.armor_max + p.structure_max > 0
              ? Math.round(((p.shield_current + p.armor_current + p.structure_current) / (p.shield_max + p.armor_max + p.structure_max)) * 100) : 0
            return (
              <div key={pi} style={{ display: 'flex', alignItems: 'center', gap: 4, padding: '1px 0', fontSize: 11 }}>
                <span className={p.is_destroyed ? 'dim' : 'blue'} style={{ minWidth: 70, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                  {p.is_destroyed ? '✗' : '▸'}{p.name}
                </span>
                {!p.is_destroyed ? (
                  <>
                    <span style={{ flex: 1 }}><HealthBar current={p.shield_current + p.armor_current + p.structure_current} max={p.shield_max + p.armor_max + p.structure_max} width={8} /></span>
                    <span className="dim" style={{ fontSize: 9, minWidth: 45, textAlign: 'right' }}>{hpPct}% {(p.distance/1000).toFixed(0)}km</span>
                  </>
                ) : <span className="dim" style={{ fontSize: 9 }}>击毁</span>}
              </div>
            )
          })}
        </div>

        {/* Enemies */}
        <div style={{ marginBottom: 4 }}>
          <div className="dim" style={{ fontSize: 10 }}>▸ 敌方 ({enemies.filter((e:any)=>!e.is_destroyed).length}/{enemies.length})</div>
          {enemies.map((e: any, i: number) => {
            const hpPct = e.shield_max + e.armor_max + e.structure_max > 0
              ? Math.round(((e.shield_current + e.armor_current + e.structure_current) / (e.shield_max + e.armor_max + e.structure_max)) * 100) : 0
            return (
              <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 4, padding: '1px 0', fontSize: 11 }}>
                <span className={e.is_destroyed ? 'dim' : 'red'} style={{ minWidth: 70, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                  {e.is_destroyed ? '✗' : '▸'}{e.name}
                </span>
                {!e.is_destroyed ? (
                  <>
                    <span style={{ flex: 1 }}><HealthBar current={e.shield_current + e.armor_current + e.structure_current} max={e.shield_max + e.armor_max + e.structure_max} width={8} /></span>
                    <span className="dim" style={{ fontSize: 9, minWidth: 45, textAlign: 'right' }}>{hpPct}% {(e.distance/1000).toFixed(0)}km</span>
                  </>
                ) : <span className="dim" style={{ fontSize: 9 }}>击毁</span>}
              </div>
            )
          })}
        </div>

        <div ref={logRef} style={{ height: 100, overflowY: 'auto', fontSize: 10, margin: '4px 0', padding: 3, background: 'var(--bg)', lineHeight: 1.4 }}>
          {logs.slice(-20).map((l: string, i: number) => (
            <div key={i} style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
              <span className={l.includes('击毁') || l.includes('通关') || l.includes('全灭') ? 'epic' : l.includes('命中') ? 'damage' : l.includes('Tick') ? 'dim' : 'info'}>{l}</span>
            </div>
          ))}
        </div>

        <div className="center">
          {!autoMode ? (
            <>
              <button className="btn" onClick={tick}>[ 下一Tick ]</button>
              <button className="btn btn-success" onClick={startAuto}>[ ▶ 自动(1秒/Tick) ]</button>
            </>
          ) : (
            <button className="btn btn-danger" onClick={stopAuto}>[ ■ 停止自动 ]</button>
          )}
          <button className="btn btn-danger" onClick={leave}>[ 撤退 ]</button>
        </div>
        {autoMode && <div className="dim center" style={{ fontSize: 11 }}>自动战斗中(WebSocket)，服务端每秒推送...</div>}
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
