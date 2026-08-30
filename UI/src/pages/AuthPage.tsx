import { useState, type FormEvent } from 'react'
import { ApiError } from '../lib/api'
import { useAuth } from '../context/AuthContext'
import { ThemeToggle } from '../components/ThemeToggle'

export function AuthPage() {
  const { login, register } = useAuth()
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')

  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    setBusy(true)
    setError('')
    try {
      if (mode === 'login') {
        await login(username.trim(), password)
      } else {
        await register(username.trim(), password)
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Something went wrong')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="auth-shell">
      <div className="auth-topbar">
        <div className="brand compact">
          <span className="brand-mark" aria-hidden>
            <span />
            <span />
            <span />
          </span>
          <h1>Recho</h1>
        </div>
        <ThemeToggle />
      </div>

      <div className="auth-card">
        <div className="auth-heading">
          <p className="eyebrow">Realtime messaging</p>
          <h2>{mode === 'login' ? 'Welcome back.' : 'Join the conversation.'}</h2>
          <p className="auth-copy">
            {mode === 'login' ? 'Sign in to continue your conversations.' : 'Create your account and start chatting.'}
          </p>
        </div>

        <div className="segmented" role="tablist" aria-label="Authentication mode">
          <button type="button" className={mode === 'login' ? 'active' : ''} onClick={() => setMode('login')}>
            Sign in
          </button>
          <button type="button" className={mode === 'register' ? 'active' : ''} onClick={() => setMode('register')}>
            Create account
          </button>
        </div>

        <form onSubmit={onSubmit} className="auth-form">
          <label>
            Username
            <input
              autoComplete="username"
              minLength={3}
              maxLength={30}
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="yourname"
              required
            />
          </label>
          <label>
            Password
            <input
              type="password"
              autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              required
            />
          </label>
          {error ? <p className="form-error">{error}</p> : null}
          <button className="primary" type="submit" disabled={busy}>
            {busy ? 'Please wait…' : mode === 'login' ? 'Sign in' : 'Create account'}
            {!busy ? <span aria-hidden>→</span> : null}
          </button>
        </form>
      </div>
    </div>
  )
}
