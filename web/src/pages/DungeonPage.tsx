import { useState, useEffect } from 'react'
import api from '../api/client'
import HealthBar from '../components/HealthBar'

type View = 'list' | 'running' | 'result'

export default function DungeonPage() {
  const [view, setView] = useState<View>('list')
  const [dungeons, setDungeons] = useState<any[]>([])
  const [raceFilter, setRaceFilter] = useState(0)
  const [waveInfo, setWaveInfo] = useState<any>(null)
  const [fighting, setFighting] = useState(false)

  const raceNames: Record<number,string> = {
    1:'白羊',2:'金牛',3:'双子',4:'巨蟹',5:'狮子',6:'处女',
    7:'天秤',8:'天蝎',9:'射手',10:'摩羯',11:'水瓶',12:'双鱼'
  }
  const diffNames = ['','巡逻','搜索','突袭','清剿','歼灭','攻坚','远征','深渊','法则','神格']

  useEffect(() => {
    checkStatus()
  }, [])

  const checkStatus = async () => {
    try {
      const res: any = await api.get('/dungeons/status')
      if (res.data) {
        setWaveInfo(res.data)
        setView('running')
      }
    } catch {
      loadDungeons()
    }
  }

  const loadDungeons = async () => {
    try {
      const q = raceFilter > 0 ? `?race=${raceFilter}` : ''
      const res: any = await api.get(`/dungeons${q}`)
      setDungeons(res.data.dungeons || [])
      setView('list')
    } catch {}
  }

  const enterDungeon = async (id: number) => {
    try {
      const res: any = await api.post('/dungeons/enter', { dungeon_def_id: id })
      setWaveInfo(res.data)
      setView('running')
    } catch (e: any) {
      alert(e?.message || '无法进入')
    }
  }

  const fightWave = async () => {
    setFighting(true)
    try {
      const res: any = await api.post('/dungeons/fight')
      setWaveInfo(res.data)
      if (res.data.combat?.status === 'completed' || res.data.combat?.status === 'finished') {
        if (res.data.wave_number >= res.data.total_waves) {
          setView('result')
        }
      }
      if (res.data.wave_text?.includes('失败')) {
        setView('result')
      }
    } catch (e: any) {
      alert(e?.message || '战斗错误')
    }
    setFighting(false)
  }

  const leaveDungeon = async () => {
    try {
      await api.post('/dungeons/leave')
      setWaveInfo(null)
      setView('list')
      loadDungeons()
    } catch {}
  }

  // ====== 远征列表 ======
  if (view === 'list') {
    return (
      <div>
        <div className="title">═══ 远征副本 ═══</div>
        <div style={{display:'flex',flexWrap:'wrap',gap:4,marginBottom:8}}>
          <button className={`btn ${raceFilter===0?'':'dim'}`} style={{fontSize:11,padding:'2px 6px'}}
            onClick={() => { setRaceFilter(0); setTimeout(loadDungeons,0) }}>全部</button>
          {Object.entries(raceNames).map(([id,name]) => (
            <button key={id} className={`btn ${raceFilter===Number(id)?'':'dim'}`}
              style={{fontSize:11,padding:'2px 6px'}}
              onClick={() => { setRaceFilter(Number(id)); setTimeout(loadDungeons,0) }}>{name}</button>
          ))}
        </div>

        {dungeons.length === 0 && <div className="dim center">加载中...</div>}
        {dungeons.map((d: any) => (
          <div key={d.id} className="item-row" onClick={() => enterDungeon(d.id)}>
            <div>
              <span className="blue">{d.name}</span>
              <div className="dim" style={{fontSize:11}}>
                难度{d.difficulty}({diffNames[d.difficulty]}) | {d.wave_count}波 | 奖励¤{(d.reward_credits/100).toLocaleString()}
              </div>
            </div>
            <span className={d.difficulty<=3?'green':d.difficulty<=6?'gold':'red'}>
              T{d.difficulty}
            </span>
          </div>
        ))}
      </div>
    )
  }

  // ====== 远征进行中 ======
  if (view === 'running' && waveInfo) {
    const combat = waveInfo.combat
    const logs = combat?.logs || []
    const participants = combat?.participants || []
    const player = participants.find((p:any) => p.type === 'player')
    const enemies = participants.filter((p:any) => p.type === 'npc')

    return (
      <div>
        <div className="title">
          ═══ 远征 第{waveInfo.wave_number}/{waveInfo.total_waves}波 ═══
          {waveInfo.is_boss && <span className="red"> [BOSS]</span>}
        </div>

        <div style={{padding:'8px 0',borderBottom:'1px dashed var(--border)'}}>
          {waveInfo.wave_text}
        </div>

        {player && (
          <div className="section">
            <div className="blue">你</div>
            <HealthBar current={player.shield_current} max={player.shield_max} label="盾" />
            <HealthBar current={player.armor_current} max={player.armor_max} label="甲" />
            <HealthBar current={player.structure_current} max={player.structure_max} label="构" />
          </div>
        )}

        {enemies.length > 0 && (
          <div className="section">
            {enemies.map((e:any,i:number) => (
              <div key={i} style={{marginBottom:4}}>
                <span className={e.is_destroyed?'dim':'red'}>{e.is_destroyed?'✗':'▸'} {e.name}</span>
                {!e.is_destroyed && (
                  <div style={{fontSize:12}}>
                    <HealthBar current={e.shield_current} max={e.shield_max} width={6} />
                  </div>
                )}
              </div>
            ))}
          </div>
        )}

        {logs.length > 0 && (
          <div className="section" style={{maxHeight:150,overflowY:'auto',fontSize:12}}>
            {logs.slice(-15).map((l:string,i:number) => (
              <div key={i} className="log-line">
                <span className={l.includes('击毁')||l.includes('通关')?'epic':l.includes('命中')?'damage':'info'}>{l}</span>
              </div>
            ))}
          </div>
        )}

        <div className="center mt">
          {waveInfo.wave_number < waveInfo.total_waves ? (
            <button className="btn btn-success" onClick={fightWave} disabled={fighting}>
              {fighting ? '战斗中...' : `[ ⚔ 战斗第${waveInfo.wave_number+1 > waveInfo.total_waves ? waveInfo.wave_number : waveInfo.wave_number}波 ]`}
            </button>
          ) : (
            <button className="btn btn-success" onClick={fightWave} disabled={fighting}>
              {fighting ? '战斗中...' : '[ ⚔ 战斗! ]'}
            </button>
          )}
          <button className="btn btn-danger" onClick={leaveDungeon}>[ 撤退 ]</button>
        </div>
      </div>
    )
  }

  // ====== 结果页 ======
  return (
    <div>
      <div className="title">═══ 远征结束 ═══</div>
      {waveInfo?.combat?.logs?.slice(-10).map((l:string,i:number) => (
        <div key={i} className="log-line">
          <span className={l.includes('通关')||l.includes('战利品')||l.includes('奖励')?'epic':l.includes('失败')?'red':'info'}>{l}</span>
        </div>
      ))}
      <div className="center mt">
        <button className="btn" onClick={() => { setWaveInfo(null); setView('list'); loadDungeons() }}>
          [ 返回远征列表 ]
        </button>
      </div>
    </div>
  )
}
