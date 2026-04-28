import { create } from 'zustand'

interface GameState {
  token: string
  charId: number
  charName: string
  raceName: string
  systemId: number
  systemName: string
  balance: number
  isLoggedIn: boolean
  login: (token: string) => void
  logout: () => void
  setChar: (id: number, name: string, race: string) => void
  setSystem: (id: number, name: string) => void
  setBalance: (b: number) => void
}

export const useGameStore = create<GameState>((set) => ({
  token: localStorage.getItem('token') || '',
  charId: parseInt(localStorage.getItem('charId') || '0'),
  charName: localStorage.getItem('charName') || '',
  raceName: '',
  systemId: 0,
  systemName: '',
  balance: 0,
  isLoggedIn: !!localStorage.getItem('token'),
  login: (token) => {
    localStorage.setItem('token', token)
    set({ token, isLoggedIn: true })
  },
  logout: () => {
    localStorage.clear()
    set({ token: '', isLoggedIn: false, charId: 0, charName: '' })
  },
  setChar: (id, name, race) => {
    localStorage.setItem('charId', String(id))
    localStorage.setItem('charName', name)
    set({ charId: id, charName: name, raceName: race })
  },
  setSystem: (id, name) => set({ systemId: id, systemName: name }),
  setBalance: (b) => set({ balance: b }),
}))
