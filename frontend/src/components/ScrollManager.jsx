import { useEffect } from 'react'
import { useLocation } from 'react-router-dom'

function ScrollManager() {
  const { pathname, hash } = useLocation()

  useEffect(() => {
    window.scrollTo({ top: 0 })
  }, [pathname])

  useEffect(() => {
    if (!hash) return
    const element = document.querySelector(hash)
    if (element) {
      element.scrollIntoView({ behavior: 'smooth', block: 'start' })
    }
  }, [hash, pathname])

  return null
}

export default ScrollManager
