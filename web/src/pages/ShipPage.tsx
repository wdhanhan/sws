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
  const [fitTarget, setFitTarget] = useState<{slotType: string, slotIndex: number} | null>(null)
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

  const loadModules = async (slotType?: string) => {
    try {
      const params = slotType ? `?category=module&slot_type=${slotType}` : '?category=module'
      const res: any = await api.get(`/items${params}`)
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
      setShowFit(false)
      setFitTarget(null)
      loadShips()
    } catch (e: any) { setMsg(e?.message || '装配失败') }
  }

  const removeModule = async (slotType: string, slotIndex: number) => {
    if (!activeShip) return
    try {
      await api.delete('/ships/fit', { data: {
        ship_id: activeShip.id, slot_type: slotType, slot_index: slotIndex
      }})
      setMsg('模块已卸载')
      loadShips()
    } catch {}
  }

  const getSlotFitting = (type: string, index: number) => {
    return fittings.find(f => f.slot_type === type && f.slot_index === index)
  }

  const fittingLabel = (f: any) => {
    if (f.damage_per_tick > 0) {
      const dps = Math.round(f.damage_per_tick / Math.max(f.rate_of_fire || 1, 1))
      const range = f.optimal_range >= 1000 ? `${(f.optimal_range/1000).toFixed(0)}km` : `${f.optimal_range}m`
      const dmgNames: Record<string,string> = {kinetic:'动',thermal:'热',em:'电',explosive:'爆'}
      return `${f.module_name} ${dps}dps ${range} ${dmgNames[f.damage_type]||''}`
    }
    if (f.bonus_type && f.bonus_value) {
      const bonusLabels: Record<string,string> = {
        shield_boost:'盾回+', shield_hp_bonus:'盾量+', armor_repair:'甲修+', armor_hp:'甲量+',
        speed_bonus:'速度+', cap_boost:'电容+', armor_kinetic_resist:'甲动抗+', armor_thermal_resist:'甲热抗+',
        armor_em_resist:'甲电抗+', armor_explosive_resist:'甲爆抗+', armor_omni_resist:'甲全抗+',
        shield_omni_resist:'盾全抗+', pg_bonus:'PG+', cpu_bonus:'CPU+', sensor_boost:'感应+',
        lock_range:'锁定距离+', microjump:'微跃',
      }
      const label = bonusLabels[f.bonus_type] || f.bonus_type+'+'
      const val = f.bonus_value < 1 ? `${Math.round(f.bonus_value*100)}%` : `${f.bonus_value}`
      return `${f.module_name} ${label}${val}`
    }
    return f.module_name
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
            <span>电容: <span className="blue">{activeShip.capacitor}</span></span>{' | '}
            <span>信号: <span className="dim">{activeShip.signature}m</span></span>
          </div>
          <div style={{fontSize:12}}>
            <span>PG: <span className="gold">{activeShip.used_pg}/{activeShip.powergrid}</span></span>{' | '}
            <span>CPU: <span className="gold">{activeShip.used_cpu}/{activeShip.cpu}</span></span>{' | '}
            <span>DPS: <span className="red">{activeShip.dps || 0}</span></span>
          </div>
          {(() => {
            const s = activeShip
            const avgShieldRes = ((s.shield_res_kinetic||0)+(s.shield_res_thermal||0)+(s.shield_res_em||0)+(s.shield_res_explosive||0))/4
            const avgArmorRes = ((s.armor_res_kinetic||0)+(s.armor_res_thermal||0)+(s.armor_res_em||0)+(s.armor_res_explosive||0))/4
            const shieldEHP = Math.round((s.shield_max||0) / Math.max(1-avgShieldRes, 0.15))
            const armorEHP = Math.round((s.armor_max||0) / Math.max(1-avgArmorRes, 0.15))
            const structEHP = Math.round((s.structure_max||0) / 0.95)
            const totalEHP = shieldEHP + armorEHP + structEHP
            const fmt = (n: number) => n >= 10000 ? `${(n/1000).toFixed(1)}k` : `${n}`
            return (
              <div style={{fontSize:12}}>
                <span>EHP: <span className="blue">{fmt(totalEHP)}</span></span>
                <span className="dim" style={{fontSize:10}}> (盾{fmt(shieldEHP)} 甲{fmt(armorEHP)} 构{fmt(structEHP)})</span>
              </div>
            )
          })()}

          <div className="section-title mt">── 抗性 ──</div>
          <div style={{fontSize:11}}>
            <div>
              <span className="dim">护盾: </span>
              <span style={{color:'#8cf'}}>动能 {Math.round((activeShip.shield_res_kinetic||0)*100)}%</span>{' '}
              <span style={{color:'#f88'}}>热能 {Math.round((activeShip.shield_res_thermal||0)*100)}%</span>{' '}
              <span style={{color:'#8ff'}}>电磁 {Math.round((activeShip.shield_res_em||0)*100)}%</span>{' '}
              <span style={{color:'#fa4'}}>爆炸 {Math.round((activeShip.shield_res_explosive||0)*100)}%</span>
            </div>
            <div>
              <span className="dim">装甲: </span>
              <span style={{color:'#8cf'}}>动能 {Math.round((activeShip.armor_res_kinetic||0)*100)}%</span>{' '}
              <span style={{color:'#f88'}}>热能 {Math.round((activeShip.armor_res_thermal||0)*100)}%</span>{' '}
              <span style={{color:'#8ff'}}>电磁 {Math.round((activeShip.armor_res_em||0)*100)}%</span>{' '}
              <span style={{color:'#fa4'}}>爆炸 {Math.round((activeShip.armor_res_explosive||0)*100)}%</span>
            </div>
          </div>

          <div className="section-title mt">── 装配 ──</div>
          {['high','mid','low'].map(slotType => {
            const label = slotType === 'high' ? '高槽' : slotType === 'mid' ? '中槽' : '低槽'
            const count = slotType === 'high' ? (activeShip.high_slots || 0)
              : slotType === 'mid' ? (activeShip.mid_slots || 0)
              : (activeShip.low_slots || 0)
            return (
              <div key={slotType} style={{marginBottom:8}}>
                <span className="dim">{label}: </span>
                {Array.from({length: count}, (_, i) => {
                  const fit = getSlotFitting(slotType, i)
                  return fit ? (
                    <span key={i} className="blue" style={{cursor:'pointer',marginRight:8,fontSize:11}}
                      onClick={() => removeModule(slotType, i)}>
                      [{fittingLabel(fit)}]
                    </span>
                  ) : (
                    <span key={i} className="dim" style={{cursor:'pointer',marginRight:8}}
                      onClick={() => { setFitTarget({slotType, slotIndex: i}); setShowFit(true); loadModules(slotType) }}>
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

      {showFit && fitTarget && (
        <div className="section mt" style={{border:'1px solid var(--border)',padding:8}}>
          <div className="section-title">── 选择模块 ({fitTarget.slotType === 'high' ? '高槽' : fitTarget.slotType === 'mid' ? '中槽' : '低槽'} #{fitTarget.slotIndex + 1}) ──</div>
          {modules.map((m: any) => {
            const dmgNames: Record<string,string> = {kinetic:'动能',thermal:'热能',em:'电磁',explosive:'爆炸'}
            const bonusLabels: Record<string,string> = {
              shield_boost:'盾回+', shield_hp_bonus:'盾量+', armor_repair:'甲修+', armor_hp:'甲量+',
              speed_bonus:'速+', cap_boost:'电容+', armor_kinetic_resist:'甲动抗+', armor_thermal_resist:'甲热抗+',
              armor_em_resist:'甲电抗+', armor_explosive_resist:'甲爆抗+', armor_omni_resist:'甲全抗+',
              shield_omni_resist:'盾全抗+', pg_bonus:'PG+', cpu_bonus:'CPU+', sensor_boost:'感应+',
              lock_range:'锁距+', microjump:'微跃',
            }
            let detail = ''
            if (m.damage_per_tick > 0) {
              const dps = Math.round(m.damage_per_tick / Math.max(m.rate_of_fire || 1, 1))
              const range = m.optimal_range >= 1000 ? `${(m.optimal_range/1000).toFixed(0)}km` : `${m.optimal_range}m`
              detail = `${dps}dps ${dmgNames[m.damage_type]||''} ${range}`
              if (m.cap_cost > 0) detail += ` 电容${m.cap_cost}`
            } else if (m.bonus_type && m.bonus_value) {
              const lbl = bonusLabels[m.bonus_type] || m.bonus_type+'+'
              const val = m.bonus_value < 1 ? `${Math.round(m.bonus_value*100)}%` : `${m.bonus_value}`
              detail = `${lbl}${val}`
            }
            return (
              <div key={m.id} className="item-row" onClick={() => fitModule(fitTarget.slotType, fitTarget.slotIndex, m.id)}>
                <span>{m.name} {detail && <span className="gold" style={{fontSize:10}}>{detail}</span>}</span>
                <span className="dim" style={{fontSize:10}}>PG:{m.pg_cost} CPU:{m.cpu_cost}</span>
              </div>
            )
          })}
          <button className="btn dim" onClick={() => { setShowFit(false); setFitTarget(null) }}>[ 取消 ]</button>
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
