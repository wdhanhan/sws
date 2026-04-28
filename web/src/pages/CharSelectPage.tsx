import { useState, useEffect } from 'react'
import api from '../api/client'
import { useGameStore } from '../stores/gameStore'

const RACE_NAMES: Record<number, string> = {
  1:'白羊座·冲锋者',2:'金牛座·铸造者',3:'双子座·幻影者',4:'巨蟹座·守护者',
  5:'狮子座·统御者',6:'处女座·精工者',7:'天秤座·裁量者',8:'天蝎座·蚀刻者',
  9:'射手座·游猎者',10:'摩羯座·筑垒者',11:'水瓶座·革新者',12:'双鱼座·共生者',
}

interface Props { onSelect: () => void }

export default function CharSelectPage({ onSelect }: Props) {
  const [chars, setChars] = useState<any[]>([])
  const [creating, setCreating] = useState(false)
  const [name, setName] = useState('')
  const [raceId, setRaceId] = useState(1)
  const setChar = useGameStore(s => s.setChar)

  const loadChars = async () => {
    try {
      const res: any = await api.get('/characters')
      setChars(res.data.characters || [])
    } catch {}
  }

  useEffect(() => { loadChars() }, [])

  const selectChar = (c: any) => {
    setChar(c.id, c.name, RACE_NAMES[c.race_id] || '')
    onSelect()
  }

  const createChar = async () => {
    try {
      await api.post('/characters', { name, race_id: raceId })
      setCreating(false)
      loadChars()
    } catch {}
  }

  return (
    <div style={{ padding: 16, maxWidth: 500, margin: '20px auto' }}>
      <div className="title">═══ 选择角色 ═══</div>

      {chars.map((c, i) => (
        <div key={c.id} className="item-row" onClick={() => selectChar(c)}>
          <div>
            <span className="blue">{i+1}. {c.name}</span>{' '}
            <span className="dim">{RACE_NAMES[c.race_id]}</span>
            <div className="dim" style={{fontSize:12}}>
              ▸ 意识完整度: {c.consciousness_pct}% | 疲劳: {c.fatigue_points}/480
            </div>
          </div>
          <span className="gold">¤{(c.balance/100).toLocaleString()}</span>
        </div>
      ))}

      {!creating ? (
        <div className="center mt">
          <button className="btn" onClick={() => setCreating(true)}>[ 创建新角色 ]</button>
        </div>
      ) : (
        <div className="mt" style={{border:'1px solid var(--border)',padding:12}}>
          <div className="section-title">── 创建角色 ──</div>
          <input className="input" placeholder="角色名(2-12字)" value={name}
            onChange={e => setName(e.target.value)} />
          <div style={{margin:'8px 0'}}>选择种族:</div>
          <div style={{display:'grid',gridTemplateColumns:'1fr 1fr',gap:4}}>
            {Object.entries(RACE_NAMES).map(([id, n]) => (
              <button key={id} className={`btn ${Number(id)===raceId?'':'dim'}`}
                style={{fontSize:12,padding:'4px 8px'}}
                onClick={() => setRaceId(Number(id))}>
                {Number(id)===raceId?'★ ':''}{n}
              </button>
            ))}
          </div>
          <div className="center mt">
            <button className="btn btn-success" onClick={createChar}>[ 确认创建 ]</button>
            <button className="btn dim" onClick={() => setCreating(false)}>[ 取消 ]</button>
          </div>
        </div>
      )}
    </div>
  )
}
