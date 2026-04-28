import { useState } from 'react'
import { useGameStore } from './stores/gameStore'
import StatusBar from './components/StatusBar'
import TabBar from './components/TabBar'
import LoginPage from './pages/LoginPage'
import CharSelectPage from './pages/CharSelectPage'
import StarMapPage from './pages/StarMapPage'
import ShipPage from './pages/ShipPage'
import CombatPage from './pages/CombatPage'
import IndustryPage from './pages/IndustryPage'
import DungeonPage from './pages/DungeonPage'
import MorePage from './pages/MorePage'

type AppState = 'login' | 'charSelect' | 'game'

export default function App() {
  const { isLoggedIn, charId } = useGameStore()
  const [appState, setAppState] = useState<AppState>(
    isLoggedIn ? (charId > 0 ? 'game' : 'charSelect') : 'login'
  )
  const [activeTab, setActiveTab] = useState('starmap')

  if (appState === 'login') {
    return (
      <div className="app">
        <LoginPage onLogin={() => setAppState('charSelect')} />
      </div>
    )
  }

  if (appState === 'charSelect') {
    return (
      <div className="app">
        <CharSelectPage onSelect={() => setAppState('game')} />
      </div>
    )
  }

  const renderPage = () => {
    switch (activeTab) {
      case 'starmap': return <StarMapPage />
      case 'ship': return <ShipPage />
      case 'combat': return <CombatPage />
      case 'industry': return <IndustryPage />
      case 'more': return <MorePage />
      default: return <StarMapPage />
    }
  }

  return (
    <div className="app">
      <StatusBar />
      <div className="main-content">
        {renderPage()}
      </div>
      <TabBar activeTab={activeTab} onTabChange={setActiveTab} />
    </div>
  )
}
