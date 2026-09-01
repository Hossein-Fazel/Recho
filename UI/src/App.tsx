import { AuthProvider, useAuth } from './context/AuthContext'
import { AuthPage } from './pages/AuthPage'
import { ChatPage } from './pages/ChatPage'

function Gate() {
  const { user, loading } = useAuth()

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

  return user ? <ChatPage /> : <AuthPage />
}

export default function App() {
  return (
    <AuthProvider>
      <Gate />
    </AuthProvider>
  )
}
