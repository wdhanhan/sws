import { useState } from 'react'
import api from '../api/client'
import { useGameStore } from '../stores/gameStore'

interface Props { onLogin: () => void }

export default function LoginPage({ onLogin }: Props) {
  const [account, setAccount] = useState('')
  const [error, setError] = useState('')
  const login = useGameStore(s => s.login)

  const handleSubmit = async () => {
    setError('')
    const name = account.trim()
    if (!name) {
      setError('请输入账号')
      return
    }
    try {
      const res: any = await api.post('/login', { account: name })
      login(res.data.access_token)
      onLogin()
    } catch (e: any) {
      setError(e?.message || '登录失败')
    }
  }

  return (
    <div style={{ padding: 20, maxWidth: 400, margin: '60px auto' }}>
      <div className="title" style={{ fontSize: 20, marginBottom: 30 }}>
        ═══════════════════<br />
        ★ 星 陨 战 歌 ★<br />
        STARFALL WARSONG<br />
        ═══════════════════
      </div>

      <input className="input" placeholder="账号 (2-20字)" value={account}
        onChange={e => setAccount(e.target.value)} maxLength={20}
        onKeyDown={e => e.key === 'Enter' && handleSubmit()} />

      {error && <div className="red" style={{ padding: 8 }}>▸ {error}</div>}

      <div className="center mt">
        <button className="btn" onClick={handleSubmit}>
          [ 进入游戏 ]
        </button>
      </div>

      <div className="dim center mt" style={{ fontSize: 12 }}>
        新账号将自动注册
      </div>
    </div>
  )
}
