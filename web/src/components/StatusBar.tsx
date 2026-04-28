import { useEffect } from 'react'
import { useGameStore } from '../stores/gameStore'
import api from '../api/client'

export default function StatusBar() {
  const { charId, charName, raceName, systemId, systemName, balance, setSystem, setBalance } = useGameStore()

  useEffect(() => {
    if (!charId) return
    const load = async () => {
      try {
        const charRes: any = await api.get(`/characters/${charId}`)
        const c = charRes.data
        if (c) {
          setBalance(c.balance)
          if (c.current_system_id && c.current_system_id !== systemId) {
            const sysRes: any = await api.get(`/starmap/systems/${c.current_system_id}`)
            if (sysRes.data) {
              setSystem(c.current_system_id, sysRes.data.system.name)
            }
          }
        }
      } catch {}
    }
    load()
    const timer = setInterval(load, 10000) // 每10秒刷新余额
    return () => clearInterval(timer)
  }, [charId])

  return (
    <div className="status-bar">
      <span className="name">{charName || '未选择'}</span>
      <span className="credits">¤ {(balance / 100).toLocaleString()}</span>
      <span>☆ {systemName || '---'}</span>
      <span className="green">● 在线</span>
    </div>
  )
}
