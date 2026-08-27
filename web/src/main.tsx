import { createRoot } from 'react-dom/client'
import './styles.css'

function App() {
  return (
    <main className="app-shell">
      <header>
        <p className="eyebrow">rePugnant</p>
        <h1>Documentation that stays close to code.</h1>
        <p className="lede">The documentation workspace will appear here once connected to a rePugnant server.</p>
      </header>
    </main>
  )
}

createRoot(document.getElementById('root')!).render(<App />)
