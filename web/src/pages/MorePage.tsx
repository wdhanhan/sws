import { useState, useEffect } from 'react'
import api from '../api/client'
import { useGameStore } from '../stores/gameStore'

type SubPage = 'menu'|'skills'|'corp'|'mail'|'chat'|'encounter'|'dungeon'|'settings'

export default function MorePage() {
  const [sub, setSub] = useState<SubPage>('menu')
  const [data, setData] = useState<any>(null)
  const logout = useGameStore(s => s.logout)

  const load = async (page: SubPage) => {
    setSub(page)
    try {
      switch(page) {
        case 'skills':
          const sk: any = await api.get('/skills')
          setData(sk.data)
          break
        case 'corp':
          const cp: any = await api.get('/corps/mine')
          setData(cp.data)
          break
        case 'mail':
          const ml: any = await api.get('/mail')
          setData(ml.data)
          break
        case 'chat':
          const ch: any = await api.get('/chat/messages?channel=trade&limit=20')
          setData(ch.data)
          break
        case 'encounter':
          const en: any = await api.post('/encounters/try')
          setData(en.data)
          break
        case 'dungeon':
          const dg: any = await api.get('/dungeons')
          setData(dg.data)
          break
      }
    } catch (e: any) {
      setData({ error: e?.message || '加载失败' })
    }
  }

  if (sub === 'menu') {
    return (
      <div>
        <div className="title">═══ 更多 ═══</div>
        {[
          { id: 'skills' as SubPage, label: '技能训练', icon: '📖' },
          { id: 'corp' as SubPage, label: '军团', icon: '⚑' },
          { id: 'mail' as SubPage, label: '邮件', icon: '✉' },
          { id: 'chat' as SubPage, label: '聊天', icon: '💬' },
          { id: 'encounter' as SubPage, label: '探索奇遇', icon: '✦' },
          { id: 'dungeon' as SubPage, label: '远征副本', icon: '⚑' },
          { id: 'settings' as SubPage, label: '设置', icon: '⚙' },
        ].map(item => (
          <div key={item.id} className="item-row" onClick={() => load(item.id)}>
            <span>{item.icon} {item.label}</span>
            <span className="dim">&gt;</span>
          </div>
        ))}
      </div>
    )
  }

  return (
    <div>
      <div className="title" style={{cursor:'pointer'}} onClick={() => setSub('menu')}>← 返回更多</div>

      {sub === 'skills' && (
        <div>
          <div className="section-title">── 已学技能 ──</div>
          {data?.completed?.length > 0 && data.completed.map((c:string,i:number) => (
            <div key={i} className="log-line epic">★ {c}</div>
          ))}
          {data?.skills?.map((s:any) => (
            <div key={s.skill_def_id} className="item-row">
              <span>[{s.category}] {s.name}</span>
              <span className="blue">Lv.{s.level}</span>
            </div>
          ))}
          {(!data?.skills || data.skills.length === 0) && <div className="dim">还没有学习任何技能</div>}
          <div className="dim mt center">技能训练请通过API: POST /skills/train</div>
        </div>
      )}

      {sub === 'corp' && (
        <div>
          <div className="section-title">── 军团信息 ──</div>
          {data?.error ? (
            <div className="dim">{data.error}</div>
          ) : data?.corp ? (
            <div>
              <div>▸ 名称: <span className="blue">{data.corp.name}</span> [{data.corp.ticker}]</div>
              <div>▸ 成员: {data.corp.member_count}人</div>
              <div>▸ 税率: {(data.corp.tax_rate*100).toFixed(0)}%</div>
              <div>▸ 你的职位: {data.your_role}</div>
            </div>
          ) : <div className="dim">未加入军团</div>}
        </div>
      )}

      {sub === 'mail' && (
        <div>
          <div className="section-title">── 邮件 ({data?.count || 0}) ──</div>
          {data?.mails?.map((m:any) => (
            <div key={m.id} className="item-row">
              <span>{m.is_read ? '' : '● '}{m.subject}</span>
              <span className="dim">{m.from_name}</span>
            </div>
          ))}
          {(!data?.mails || data.mails.length === 0) && <div className="dim">没有邮件</div>}
        </div>
      )}

      {sub === 'chat' && (
        <div>
          <div className="section-title">── 聊天 ──</div>
          {data?.messages?.map((m:any) => (
            <div key={m.id} className="log-line">
              <span className="blue">{m.sender_name}</span>: {m.content}
            </div>
          ))}
          {(!data?.messages || data.messages.length === 0) && <div className="dim">暂无消息</div>}
        </div>
      )}

      {sub === 'encounter' && (
        <div>
          <div className="section-title">── 奇遇 ──</div>
          {data?.triggered ? (
            <div>
              <div className="gold">★ {data.event.name}</div>
              <div className="mt" style={{whiteSpace:'pre-wrap'}}>{data.event.intro_text}</div>
              <div className="mt">
                {data.event.choices?.map((c:any) => (
                  <button key={c.index} className="btn" style={{display:'block',width:'100%',textAlign:'left',marginBottom:4}}
                    onClick={async () => {
                      try {
                        const res: any = await api.post('/encounters/choose', { encounter_id: data.event.id, choice_index: c.index })
                        setData({ result: res.data })
                      } catch {}
                    }}>
                    [{c.index}] {c.text}
                  </button>
                ))}
              </div>
            </div>
          ) : data?.result ? (
            <div>
              <div style={{whiteSpace:'pre-wrap'}}>{data.result.result_text}</div>
              {data.result.messages?.map((m:string,i:number) => (
                <div key={i} className="log-line epic">▸ {m}</div>
              ))}
            </div>
          ) : (
            <div className="dim">{data?.message || '没有发现奇遇事件。继续探索吧。'}</div>
          )}
        </div>
      )}

      {sub === 'dungeon' && (
        <div>
          <div className="section-title">── 远征副本 ──</div>
          <div className="dim mb">远征入口在战斗地点中随机刷出，难度T6+</div>
          {data?.dungeons?.slice(0, 15).map((d: any) => (
            <div key={d.id} className="item-row">
              <span>{d.name}</span>
              <span className="dim">T{d.difficulty} {d.wave_count}波 ¤{(d.reward_credits/100).toLocaleString()}</span>
            </div>
          ))}
          {!data?.dungeons && <div className="dim">加载中...</div>}
        </div>
      )}

      {sub === 'settings' && (
        <div>
          <div className="section-title">── 设置 ──</div>
          <div className="center mt">
            <button className="btn btn-danger" onClick={() => { logout(); window.location.reload() }}>
              [ 退出登录 ]
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
