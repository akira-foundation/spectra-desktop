import { Moon, Sun } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useTheme } from '@/hooks/useTheme'

export function ThemeSwitcher() {
  const { theme, setTheme } = useTheme()

  const toggleTheme = () => {
    if (theme === 'dark') {
      setTheme('light')
    } else if (theme === 'light') {
      setTheme('dark')
    } else {
      // If system, switch to opposite of current appearance
      const isDark = document.documentElement.classList.contains('dark')
      setTheme(isDark ? 'light' : 'dark')
    }
  }

  const isDark = theme === 'dark' || (theme === 'system' && document.documentElement.classList.contains('dark'))

  return (
    <Button
      variant="ghost"
      size="icon-sm"
      onClick={toggleTheme}
      title={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
      className="h-7 w-7 text-white/85 hover:bg-white/10 hover:text-white"
    >
      {isDark ? <Sun className="w-3.5 h-3.5" /> : <Moon className="w-3.5 h-3.5" />}
    </Button>
  )
}
