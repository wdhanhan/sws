import { useState } from 'react'
import api from '../api/client'
import { useGameStore } from '../stores/gameStore'

interface Props { onLogin: () => void }

export default function LoginPage({ onLogin }: Props) {
  const [phone, setPhone] = useState('')
  const [password, setPassword] = useState('')
  const [isRegister, setIsRegister] = useState(false)
  const [error, setError] = useState('')
  const login = useGameStore(s => s.login)

  const handleSubmit = async () => {
    setError('')
    try {
      const endpoint = isRegister ? '/register' : '/login'
      const res: any = await api.post(endpoint, { phone, password })
      login(res.data.access_token)
      onLogin()
    } catch (e: any) {
      setError(e?.message || '操作失败')
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

      <input className="input" placeholder="手机号" value={phone}
        onChange={e => setPhone(e.target.value)} />
      <input className="input" placeholder="密码(至少8位)" type="password"
        value={password} onChange={e => setPassword(e.target.value)}
        onKeyDown={e => e.key === 'Enter' && handleSubmit()} />

      {error && <div className="red" style={{ padding: 8 }}>▸ {error}</div>}

      <div className="center mt">
        <button className="btn" onClick={handleSubmit}>
          {isRegister ? '[ 注  册 ]' : '[ 登  录 ]'}
        </button>
        <button className="btn dim" onClick={() => setIsRegister(!isRegister)}>
          {isRegister ? '已有账号？登录' : '没有账号？注册'}
        </button>
      </div>

      <div className="dim center mt" style={{ fontSize: 12 }}>
        星陨纪元·新纪元的开拓者
      </div>
    </div>
  )
}
