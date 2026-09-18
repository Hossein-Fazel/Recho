import { useState } from 'react'
import { AuthProvider, useAuth } from './context/AuthContext'
import { useAppHeight } from './hooks/useAppHeight'
import { inviteCodeFromPath } from './lib/invite'
import { AuthPage } from './pages/AuthPage'
import { ChatPage } from './pages/ChatPage'

function Gate() {
  const { user, loading } = useAuth()

  // Read the invite code once on mount. Kept in state so it survives the
  // login step when someone opens an invite link while signed out.
  const [inviteCode] = useState(() =>
    inviteCodeFromPath(window.location.pathname),
  )

  if (loading) {
    return (
      <div className="splash">
        <span className="brand-mark lg" aria-hidden>
          <img src="/logo.png" alt="Recho" />
        </span>
        <p>Recho</p>
      </div>
    )
  }

  return user ? <ChatPage inviteCode={inviteCode} /> : <AuthPage />
}

export default function App() {
  useAppHeight()

  return (
    <AuthProvider>
      <Gate />
    </AuthProvider>
  )
}
