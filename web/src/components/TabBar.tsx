interface Props {
  activeTab: string
  onTabChange: (tab: string) => void
}

const tabs = [
  { id: 'starmap', label: '☆ 星图' },
  { id: 'ship', label: '⛵ 飞船' },
  { id: 'combat', label: '⚔ 战斗' },
  { id: 'industry', label: '⛏ 工业' },
  { id: 'more', label: '☰ 更多' },
]

export default function TabBar({ activeTab, onTabChange }: Props) {
  return (
    <div className="tab-bar">
      {tabs.map(t => (
        <button
          key={t.id}
          className={activeTab === t.id ? 'active' : ''}
          onClick={() => onTabChange(t.id)}
        >
          {t.label}
        </button>
      ))}
    </div>
  )
}
