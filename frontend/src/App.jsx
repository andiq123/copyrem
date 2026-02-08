import { useState, useEffect } from 'react'
import ConvertPage from './features/convert/ConvertPage'
import SeparatePage from './features/separate/SeparatePage'

const PAGES = { convert: 'convert', separate: 'separate' }

function getPageFromHash() {
  const hash = window.location.hash.slice(1)
  return hash === 'separate' ? PAGES.separate : PAGES.convert
}

function syncHash(page) {
  const want = page === PAGES.separate ? '#separate' : '#convert'
  if (window.location.hash !== want) {
    window.location.hash = want
  }
}

export default function App() {
  const [page, setPage] = useState(getPageFromHash)

  useEffect(() => {
    const onHashChange = () => setPage(getPageFromHash())
    window.addEventListener('hashchange', onHashChange)
    return () => window.removeEventListener('hashchange', onHashChange)
  }, [])

  useEffect(() => syncHash(page), [page])

  return (
    <div className="app-shell">
      <nav className="app-nav">
        <a
          href="#convert"
          className={page === PAGES.convert ? 'active' : ''}
          onClick={(e) => { e.preventDefault(); setPage(PAGES.convert) }}
        >
          Convert
        </a>
        <a
          href="#separate"
          className={page === PAGES.separate ? 'active' : ''}
          onClick={(e) => { e.preventDefault(); setPage(PAGES.separate) }}
        >
          Separate
        </a>
      </nav>
      {page === PAGES.convert ? <ConvertPage /> : <SeparatePage />}
    </div>
  )
}
