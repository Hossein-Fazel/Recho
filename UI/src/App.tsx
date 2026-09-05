import { useState } from 'react'
import { AuthProvider, useAuth } from './context/AuthContext'
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
          <span />
          <span />
          <span />
        </span>
        <p>Recho</p>
      </div>
    )
  }

  return user ? <ChatPage inviteCode={inviteCode} /> : <AuthPage />
}

export default function App() {
  return (
    <AuthProvider>
      <Gate />
    </AuthProvider>
  )
}
