import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'

const ThemeContext = createContext<{ theme: 'light' | 'dark'; setTheme: (theme: 'light' | 'dark') => void } | null>(null)

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setTheme] = useState<'light' | 'dark'>(() => localStorage.getItem('theme') === 'dark' ? 'dark' : 'light')
  useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark')
    localStorage.setItem('theme', theme)
  }, [theme])
  return <ThemeContext.Provider value={{ theme, setTheme }}>{children}</ThemeContext.Provider>
}

export function useTheme() {
  const theme = useContext(ThemeContext)
  if (!theme) throw new Error('ThemeProvider required')
  return theme
}
