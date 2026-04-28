import { useState } from 'react'
import api from '../api/client'

type SubPage = 'menu' | 'mining' | 'refine' | 'market' | 'cargo' | 'assets'

export default function IndustryPage() {
  const [sub, setSub] = useState<SubPage>('menu')
  const [miningStatus, setMiningStatus] = useState<any>(null)
  const [inventory, setInventory] = useState<any[]>([])
  const [recipes, setRecipes] = useState<any[]>([])
  const [orders, setOrders] = useState<any[]>([])
  const [assets, setAssets] = useState<any>({})
  const [msg, setMsg] = useState('')

  const startMining = async (beltId: number) => {
    try {
      const res: any = await api.post('/mining/start', { belt_id: beltId })
      setMiningStatus(res.data)
      setMsg('采矿开始')
    } catch (e: any) { setMsg(e?.message || '失败') }
  }

  const collectMining = async () => {
    try {
      const res: any = await api.post('/mining/collect')
      setMiningStatus(res.data)
      setMsg(`收集了 ${res.data.total_mined} 单位矿石`)
    } catch (e: any) { setMsg(e?.message || '失败') }
  }

  const stopMining = async () => {
    try { await api.post('/mining/stop'); setMiningStatus(null); setMsg('采矿停止') }
    catch (e: any) { setMsg(e?.message || '失败') }
  }

  const loadInventory = async () => {
    try {
      // 先获取角色当前星系
      const charId = localStorage.getItem('charId')
      let systemId = 0
      if (charId) {
        const charRes: any = await api.get(`/characters/${charId}`)
        systemId = charRes.data?.current_system_id || 0
      }
      const res: any = await api.get(`/inventory?system_id=${systemId}`)
      setInventory(res.data.items || [])
    } catch {}
  }

  const loadRecipes = async () => {
    try {
      const res: any = await api.get('/refine/recipes')
      setRecipes(res.data.recipes || [])
    } catch {}
  }

  const loadAssets = async () => {
    try {
      const res: any = await api.get('/inventory/assets')
      setAssets(res.data.assets || {})
    } catch {}
  }

  const transferItem = async (itemDefId: number, qty: number, direction: string) => {
    try {
      const res: any = await api.post('/inventory/transfer', { item_def_id: itemDefId, quantity: qty, direction })
      setMsg(res.data.message || '转移成功')
      loadInventory()
    } catch (e: any) { setMsg(e?.message || '转移失败') }
  }

  const loadOrders = async () => {
    try {
      const res: any = await api.get('/market/orders?limit=20')
      setOrders(res.data.orders || [])
    } catch {}
  }

  if (sub === 'menu') {
    return (
      <div>
        <div className="title">═══ 工业中心 ═══</div>
        {[
          { id: 'mining' as SubPage, icon: '⛏', label: '采矿', desc: '开采小行星矿石' },
          { id: 'refine' as SubPage, icon: '⚗', label: '精炼', desc: '矿石→金属' },
          { id: 'market' as SubPage, icon: '¤', label: '市场', desc: '买卖物品' },
          { id: 'cargo' as SubPage, icon: '◆', label: '舰船货仓', desc: '当前舰船中的物品' },
          { id: 'assets' as SubPage, icon: '☰', label: '资产总览', desc: '所有位置的物品' },
        ].map(item => (
          <div key={item.id} className="item-row" onClick={() => { setSub(item.id); if(item.id==='cargo')loadInventory(); if(item.id==='refine')loadRecipes(); if(item.id==='market')loadOrders(); if(item.id==='assets')loadAssets() }}>
            <span>{item.icon} {item.label}</span>
            <span className="dim">{item.desc}</span>
          </div>
        ))}
      </div>
    )
  }

  return (
    <div>
      <div className="title" style={{cursor:'pointer'}} onClick={() => setSub('menu')}>← 返回工业中心</div>
      {msg && <div className="gold" style={{padding:4}}>▸ {msg}</div>}

      {sub === 'mining' && (
        <div>
          <div className="section-title">── 采矿 ──</div>
          {miningStatus?.is_mining ? (
            <div>
              <div>▸ 矿石: {miningStatus.ore_name}</div>
              <div>▸ 产量: {miningStatus.yield_per_cycle}/周期 ({miningStatus.cycle_time_sec}秒)</div>
              <div className="center mt">
                <button className="btn btn-success" onClick={collectMining}>[ 收集 ]</button>
                <button className="btn btn-danger" onClick={stopMining}>[ 停止 ]</button>
              </div>
            </div>
          ) : (
            <div>
              <div className="dim">选择矿带开始采矿 (输入矿带ID):</div>
              {[1,2,3].map(id => (
                <button key={id} className="btn" onClick={() => startMining(id)}>[ 矿带 {id} ]</button>
              ))}
            </div>
          )}
        </div>
      )}

      {sub === 'refine' && (
        <div>
          <div className="section-title">── 精炼配方 ──</div>
          {recipes.map((r: any, i: number) => (
            <div key={i} className="item-row">
              <span>{r.input} x{r.input_quantity}</span>
              <span className="blue">→ {r.output} x{r.output_quantity}</span>
            </div>
          ))}
        </div>
      )}

      {sub === 'market' && (
        <div>
          <div className="section-title">── 市场订单 ──</div>
          {orders.length === 0 && <div className="dim">暂无活跃订单</div>}
          {orders.map((o: any) => (
            <div key={o.id} className="item-row">
              <span className={o.order_type==='sell'?'red':'green'}>[{o.order_type==='sell'?'卖':'买'}] 物品#{o.item_def_id}</span>
              <span>¤{o.price} x{o.quantity}</span>
            </div>
          ))}
        </div>
      )}

      {sub === 'cargo' && (
        <div>
          <div className="section-title">── 舰船货仓 ──</div>
          {inventory.length === 0 && <div className="dim">货仓为空</div>}
          {inventory.map((item: any) => (
            <div key={item.id} className="item-row">
              <div>
                <span>◆ {item.name}</span>
                <span className="gold"> x{item.quantity}</span>
              </div>
              <button className="btn dim" style={{fontSize:10,padding:'1px 6px'}}
                onClick={(e) => { e.stopPropagation(); transferItem(item.item_def_id, item.quantity, 'to_station') }}>
                → 存入空间站
              </button>
            </div>
          ))}
        </div>
      )}

      {sub === 'assets' && (
        <div>
          <div className="section-title">── 资产总览 ──</div>
          {Object.keys(assets).length === 0 && <div className="dim">暂无资产</div>}
          {Object.entries(assets).map(([location, items]: [string, any]) => (
            <div key={location} style={{marginBottom:12}}>
              <div className="blue" style={{fontSize:12,borderBottom:'1px dashed var(--border)',padding:'4px 0'}}>{location}</div>
              {(items as any[]).map((item: any, i: number) => (
                <div key={i} className="item-row" style={{paddingLeft:12}}>
                  <span>◆ {item.name}</span>
                  <span className="gold">x{item.quantity}</span>
                </div>
              ))}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
