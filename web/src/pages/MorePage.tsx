import { useState } from 'react'
import api from '../api/client'
import { useGameStore } from '../stores/gameStore'

type SubPage = 'menu'|'skills'|'corp'|'mail'|'chat'|'encounter'|'dungeon'|'fleet'|'wrecks'|'settings'

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
        case 'fleet':
          const fl: any = await api.get('/fleet')
          setData(fl.data)
          break
        case 'wrecks':
          const wr: any = await api.get('/wrecks')
          setData(wr.data)
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
          { id: 'fleet' as SubPage, label: '舰队', icon: '⚓' },
          { id: 'wrecks' as SubPage, label: '残骸打捞', icon: '♻' },
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

      {sub === 'wrecks' && (
        <div>
          <div className="section-title">── 残骸打捞 ──</div>
          {data?.wrecks?.length > 0 ? data.wrecks.map((w: any) => (
            <div key={w.id} className="item-row" style={{flexDirection:'column',alignItems:'stretch',gap:2}}>
              <div style={{display:'flex',justifyContent:'space-between'}}>
                <span>
                  <span className="gold">⚠</span> {w.ship_type || '未知舰船'} <span className="dim">"{w.ship_name}"</span>
                  <span className="dim" style={{fontSize:10}}> [{w.owner_name}]</span>
                </span>
                <span className="dim" style={{fontSize:10}}>{w.item_count}件物品</span>
              </div>
              <button className="btn btn-success" style={{fontSize:10,padding:'2px 8px'}} onClick={async () => {
                try {
                  const res: any = await api.post(`/wrecks/${w.id}/loot`)
                  const items = res.data?.items || []
                  alert(`拾取成功! 获得 ${items.length} 件物品:\n${items.map((i:any)=>`${i.name} x${i.quantity}`).join('\n')}`)
                  load('wrecks')
                } catch (e: any) { alert(e?.message || '拾取失败') }
              }}>[ 拾取残骸 ]</button>
            </div>
          )) : <div className="dim center">当前星系没有残骸</div>}
          <button className="btn dim mt" onClick={() => load('wrecks')}>[ 刷新 ]</button>
        </div>
      )}

      {sub === 'fleet' && (
        <FleetPanel data={data} reload={() => load('fleet')} />
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

function FleetPanel({ data, reload }: { data: any, reload: () => void }) {
  const [inviteId, setInviteId] = useState('')
  const [msg, setMsg] = useState('')
  const charId = useGameStore(s => s.charId)

  const fleet = data?.fleet
  const invites = data?.invites || []
  const myChars: any[] = data?.my_chars || []

  const act = async (action: string, body?: any) => {
    try {
      const res: any = await api.post(`/fleet/${action}`, body || {})
      setMsg(res.message || '操作成功')
      setTimeout(reload, 300)
    } catch (e: any) { setMsg(e?.message || '操作失败') }
  }

  return (
    <div>
      <div className="section-title">── 舰队 ──</div>
      {msg && <div className="gold" style={{padding:4}}>▸ {msg}</div>}

      {invites.length > 0 && (
        <div className="section" style={{border:'1px solid var(--border)',padding:8,marginBottom:8}}>
          <div className="dim">收到的邀请:</div>
          {invites.map((inv: any) => (
            <div key={inv.fleet_id} className="item-row">
              <span>{inv.leader_name} 的舰队</span>
              <span>
                <button className="btn btn-success" style={{fontSize:10}} onClick={() => act('accept', {fleet_id: inv.fleet_id})}>接受</button>
                <button className="btn dim" style={{fontSize:10}} onClick={() => act('decline', {fleet_id: inv.fleet_id})}>拒绝</button>
              </span>
            </div>
          ))}
        </div>
      )}

      {fleet ? (
        <div>
          <div>▸ 舰队 #{fleet.fleet_id} | 队长: <span className="blue">{fleet.leader_name}</span></div>
          <div className="section-title mt">── 成员 ──</div>
          {fleet.members?.map((m: any) => (
            <div key={m.character_id} className="item-row">
              <span>
                {m.character_id === fleet.leader_id ? '★ ' : ''}
                <span className={m.status === 'joined' ? 'blue' : 'dim'}>{m.name}</span>
                {m.ship_name && <span className="dim" style={{fontSize:10}}> ({m.ship_name})</span>}
                {m.status === 'pending' && <span className="gold" style={{fontSize:10}}> [待接受]</span>}
              </span>
              {fleet.leader_id === charId && m.character_id !== charId && m.status === 'joined' && (
                <button className="btn dim" style={{fontSize:10}} onClick={() => act('kick', {char_id: m.character_id})}>踢出</button>
              )}
            </div>
          ))}

          {fleet.leader_id === charId && (
            <div className="mt">
              {myChars.length > 0 && (
                <div style={{marginBottom:6}}>
                  <div className="dim" style={{fontSize:10,marginBottom:4}}>快捷邀请(同账号):</div>
                  <div style={{display:'flex',flexWrap:'wrap',gap:4}}>
                    {myChars.filter(c => !fleet.members?.some((m:any) => m.character_id === c.id)).map((c: any) => (
                      <button key={c.id} className="btn" style={{fontSize:10,padding:'2px 8px'}}
                        onClick={() => act('invite', {char_id: c.id})}>
                        +{c.name}
                      </button>
                    ))}
                    {myChars.filter(c => !fleet.members?.some((m:any) => m.character_id === c.id)).length === 0 && (
                      <span className="dim" style={{fontSize:10}}>所有角色已在队中</span>
                    )}
                  </div>
                </div>
              )}
              <div style={{display:'flex',gap:4,alignItems:'center'}}>
                <input
                  style={{flex:1,background:'var(--bg)',border:'1px solid var(--border)',color:'var(--text)',padding:'4px 8px',fontSize:12}}
                  placeholder="输入角色ID邀请..."
                  value={inviteId}
                  onChange={e => setInviteId(e.target.value)}
                  onKeyDown={e => { if (e.key === 'Enter' && inviteId) { act('invite', {char_id: parseInt(inviteId)}); setInviteId('') } }}
                />
                <button className="btn" onClick={() => { if(inviteId) { act('invite', {char_id: parseInt(inviteId)}); setInviteId('') } }}>邀请</button>
              </div>
            </div>
          )}

          <div className="center mt">
            {fleet.leader_id === charId ? (
              <button className="btn btn-danger" onClick={() => act('disband')}>[ 解散舰队 ]</button>
            ) : (
              <button className="btn btn-danger" onClick={() => act('leave')}>[ 离开舰队 ]</button>
            )}
          </div>
        </div>
      ) : (
        <div>
          <div className="dim center">你不在任何舰队中</div>
          <div className="center mt">
            <button className="btn btn-success" onClick={() => act('create')}>[ 创建舰队 ]</button>
          </div>
        </div>
      )}
    </div>
  )
}
