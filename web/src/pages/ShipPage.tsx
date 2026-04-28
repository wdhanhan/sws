import { useState, useEffect } from 'react'
import api from '../api/client'
import HealthBar from '../components/HealthBar'

export default function ShipPage() {
  const [ships, setShips] = useState<any[]>([])
  const [shipDefs, setShipDefs] = useState<any[]>([])
  const [fittings, setFittings] = useState<any[]>([])
  const [activeShip, setActiveShip] = useState<any>(null)
  const [showBoard, setShowBoard] = useState(false)
  const [showFit, setShowFit] = useState(false)
  const [modules, setModules] = useState<any[]>([])
  const [msg, setMsg] = useState('')

  const loadShips = async () => {
    try {
      const res: any = await api.get('/ships')
      setShips(res.data.ships || [])
      const active = (res.data.ships || []).find((s:any) => s.is_active)
      if (active) {
        setActiveShip(active)
        loadFitting(active.id)
      }
    } catch {}
  }

  const loadFitting = async (shipId: number) => {
    try {
      const res: any = await api.get(`/ships/${shipId}/fitting`)
      setFittings(res.data.fittings || [])
    } catch {}
  }

  const loadShipDefs = async () => {
    try {
      const res: any = await api.get('/ships/defs')
      setShipDefs(res.data.ships || [])
    } catch {}
  }

  const loadModules = async () => {
    try {
      const res: any = await api.get('/items?category=module')
      setModules(res.data.items || [])
    } catch {}
  }

  useEffect(() => { loadShips() }, [])

  const boardShip = async (defId: number, defName: string) => {
    try {
      await api.post('/ships/board', { ship_def_id: defId, name: defName })
      setMsg(`已登上 ${defName}`)
      setShowBoard(false)
      loadShips()
    } catch (e: any) { setMsg(e?.message || '失败') }
  }

  const fitModule = async (slotType: string, slotIndex: number, moduleId: number) => {
    if (!activeShip) return
    try {
      await api.post('/ships/fit', {
        ship_id: activeShip.id, slot_type: slotType, slot_index: slotIndex, module_item_id: moduleId
      })
      setMsg('模块已装配')
      loadFitting(activeShip.id)
      setShowFit(false)
    } catch (e: any) { setMsg(e?.message || '装配失败') }
  }

  const removeModule = async (slotType: string, slotIndex: number) => {
    if (!activeShip) return
    try {
      await api.delete('/ships/fit', { data: {
        ship_id: activeShip.id, slot_type: slotType, slot_index: slotIndex
      }})
      setMsg('模块已卸载')
      loadFitting(activeShip.id)
    } catch {}
  }

  const getSlotFitting = (type: string, index: number) => {
    return fittings.find(f => f.slot_type === type && f.slot_index === index)
  }

  return (
    <div>
      <div className="title">═══ 舰船 ═══</div>
      {msg && <div className="gold" style={{padding:4}}>▸ {msg}</div>}

      {activeShip ? (
        <div className="section">
          <div className="section-title">── 当前舰船 ──</div>
          <div>▸ <span className="blue">{activeShip.def_name}</span> "{activeShip.name}" <span className="dim">T{activeShip.tier} {activeShip.ship_class}/{activeShip.ship_role}</span></div>
          <div><HealthBar current={activeShip.shield_current} max={activeShip.shield_max || activeShip.shield_current} label="护盾" />
            <span className="dim" style={{fontSize:11}}> ({activeShip.shield_current}/{activeShip.shield_max})</span></div>
          <div><HealthBar current={activeShip.armor_current} max={activeShip.armor_max || activeShip.armor_current} label="装甲" />
            <span className="dim" style={{fontSize:11}}> ({activeShip.armor_current}/{activeShip.armor_max})</span></div>
          <div><HealthBar current={activeShip.structure_current} max={activeShip.structure_max || activeShip.structure_current} label="结构" />
            <span className="dim" style={{fontSize:11}}> ({activeShip.structure_current}/{activeShip.structure_max})</span></div>
          <div style={{marginTop:6,fontSize:12}}>
            <span>速度: <span className="blue">{activeShip.max_speed}m/s</span></span>{' | '}
            <span>PG: <span className="gold">{activeShip.powergrid}</span></span>{' | '}
            <span>CPU: <span className="gold">{activeShip.cpu}</span></span>
          </div>

          <div className="section-title mt">── 装配 ──</div>
          {['high','mid','low'].map(slotType => {
            const label = slotType === 'high' ? '高槽' : slotType === 'mid' ? '中槽' : '低槽'
            const count = slotType === 'high' ? 4 : 3
            return (
              <div key={slotType} style={{marginBottom:8}}>
                <span className="dim">{label}: </span>
                {Array.from({length: count}, (_, i) => {
                  const fit = getSlotFitting(slotType, i)
                  return fit ? (
                    <span key={i} className="blue" style={{cursor:'pointer',marginRight:8}}
                      onClick={() => removeModule(slotType, i)}>
                      [{fit.module_name}]
                    </span>
                  ) : (
                    <span key={i} className="dim" style={{cursor:'pointer',marginRight:8}}
                      onClick={() => { setShowFit(true); loadModules() }}>
                      [空]
                    </span>
                  )
                })}
              </div>
            )
          })}
        </div>
      ) : (
        <div className="dim center">没有激活的舰船</div>
      )}

      <div className="center mt">
        <button className="btn" onClick={() => { setShowBoard(!showBoard); if(!showBoard) loadShipDefs() }}>
          [ {showBoard ? '取消' : '登上新船'} ]
        </button>
      </div>

      {showBoard && (
        <div className="section mt">
          <div className="section-title">── 可用舰船型号 ──</div>
          {shipDefs.slice(0, 20).map((d: any) => (
            <div key={d.id} className="item-row" onClick={() => boardShip(d.id, d.name)}>
              <span>T{d.tier} {d.name} ({d.ship_class}/{d.ship_role})</span>
              <span className="dim">护盾{d.shield_hp} 装甲{d.armor_hp}</span>
            </div>
          ))}
        </div>
      )}

      {showFit && (
        <div className="section mt" style={{border:'1px solid var(--border)',padding:8}}>
          <div className="section-title">── 选择模块 ──</div>
          {modules.map((m: any) => (
            <div key={m.id} className="item-row" onClick={() => fitModule('high', 0, m.id)}>
              <span>{m.name}</span>
              <span className="dim">PG:{m.pg_cost} CPU:{m.cpu_cost}</span>
            </div>
          ))}
          <button className="btn dim" onClick={() => setShowFit(false)}>[ 取消 ]</button>
        </div>
      )}

      {ships.length > 1 && (
        <div className="section mt">
          <div className="section-title">── 其他舰船 ──</div>
          {ships.filter(s => !s.is_active).map(s => (
            <div key={s.id} className="item-row dim">
              {s.def_name || `舰船#${s.ship_def_id}`} - {s.name}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
