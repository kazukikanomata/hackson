import { useState } from 'react'
import { Button } from '@/components/ui/button'

function App() {
  const [message, setMessage] = useState<string>('')

  const fetchHello = async () => {
    const res = await fetch('/api/hello')
    const data = (await res.json()) as { message: string }
    setMessage(data.message)
  }

  return (
    <main className="flex min-h-svh flex-col items-center justify-center gap-6">
      <h1 className="text-3xl font-bold tracking-tight">Hackathon Starter</h1>
      <Button onClick={fetchHello}>Call Go API</Button>
      {message && <p className="text-muted-foreground">{message}</p>}
    </main>
  )
}

export default App
