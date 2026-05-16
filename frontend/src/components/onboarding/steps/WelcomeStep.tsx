import { useEffect, useState } from 'react'

export function WelcomeStep() {
  const [visible, setVisible] = useState(false)
  useEffect(() => {
    const t = setTimeout(() => setVisible(true), 50)
    return () => clearTimeout(t)
  }, [])

  return (
    <div
      className={`flex flex-col items-center text-center transition-all duration-700 ease-out ${
        visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-2'
      }`}
    >
      <div className="h-20 w-20 mb-8">
        <img
          src="/favicon-light.svg"
          alt="Spectra"
          className="h-20 w-20 rounded-2xl shadow-2xl shadow-primary/20 dark:hidden"
        />
        <img
          src="/favicon.svg"
          alt="Spectra"
          className="h-20 w-20 rounded-2xl shadow-2xl shadow-primary/30 hidden dark:block"
        />
      </div>

      <h1 className="text-[44px] font-semibold tracking-tight leading-none">
        Welcome to Spectra
      </h1>

      <p className="mt-4 text-[15px] text-muted-foreground max-w-sm leading-relaxed">
        A local-first API inspector. Drop a backend folder, see every route,
        run requests, mock responses. Everything stays on your machine.
      </p>
    </div>
  )
}
