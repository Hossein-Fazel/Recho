import { useEffect, useMemo, useState } from 'react'
import { Avatar } from './Avatar'
import { ThemeToggle } from './ThemeToggle'
import { displayName, formatTime } from '../lib/format'
import { api } from '../lib/api'
import type { Conversation, User, UserSearch } from '../lib/types'

type SidebarProps = {
  user: User
  conversations: Conversation[]
  activeId: string | null
  onSelect: (conversation: Conversation) => void
  onCreated: (conversationId: string, peer: UserSearch) => void
  onLogout: () => void
  onCloseMobile?: () => void
}

export function Sidebar({
  user,
  conversations,
  activeId,
  onSelect,
  onCreated,
  onLogout,
  onCloseMobile,
}: SidebarProps) {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<UserSearch[]>([])
  const [searching, setSearching] = useState(false)

  useEffect(() => {
    const q = query.trim()
    if (q.length < 1) {
      setResults([])
      return
    }

    const handle = window.setTimeout(async () => {
      setSearching(true)
      try {
        const res = await api.searchUsers(q)
        setResults((res.users ?? []).filter((u) => u.id !== user.id))
      } catch {
        setResults([])
      } finally {
        setSearching(false)
      }
    }, 220)

    return () => window.clearTimeout(handle)
  }, [query, user.id])

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q || results.length > 0) return conversations
    return conversations.filter((c) => displayName(c).toLowerCase().includes(q))
  }, [conversations, query, results.length])

  async function startChat(peer: UserSearch) {
    const res = await api.openDirect(peer.id)
    onCreated(res.conversation_id, peer)
    setQuery('')
    setResults([])
    onCloseMobile?.()
  }

  return (
    <aside className="sidebar">
      <header className="sidebar-head">
        <div className="brand compact">
          <span className="brand-mark" aria-hidden>
            <span />
            <span />
            <span />
          </span>
          <div>
            <h1>Recho</h1>
            <p className="eyebrow">Messages</p>
          </div>
        </div>
        <ThemeToggle />
      </header>

      <div className="search-wrap">
        <div className="search-box">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <circle cx="11" cy="11" r="6.5" />
            <path d="m16 16 4 4" />
          </svg>
          <input
            className="search"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search"
            aria-label="Search people or conversations"
          />
          {query ? (
            <button type="button" className="clear-search" onClick={() => setQuery('')} aria-label="Clear search">
              ×
            </button>
          ) : null}
        </div>
      </div>

      {results.length > 0 ? (
        <div className="result-block">
          <p className="list-label">People</p>
          <ul className="conv-list">
            {results.map((peer) => (
              <li key={peer.id}>
                <button type="button" className="conv-item" onClick={() => void startChat(peer)}>
                  <Avatar id={peer.id} name={displayName(peer)} url={peer.avatar_url} />
                  <span className="conv-meta">
                    <span className="conv-name">{displayName(peer)}</span>
                    <span className="conv-preview">@{peer.username}</span>
                  </span>
                  <span className="new-chat-arrow" aria-hidden>→</span>
                </button>
              </li>
            ))}
          </ul>
        </div>
      ) : null}

      {searching ? <p className="muted pad">Searching…</p> : null}

      <p className="list-label">Conversations</p>
      <ul className="conv-list grow">
        {filtered.length === 0 ? (
          <li className="empty-list">
            <span className="empty-list-icon">⌁</span>
            <span>{query ? 'No matches found' : 'No conversations yet'}</span>
            {!query ? <small>Search for someone to start chatting.</small> : null}
          </li>
        ) : (
          filtered.map((conv) => {
            const name = displayName(conv)
            return (
              <li key={conv.conversation_id}>
                <button
                  type="button"
                  className={`conv-item ${activeId === conv.conversation_id ? 'active' : ''}`}
                  onClick={() => {
                    onSelect(conv)
                    onCloseMobile?.()
                  }}
                >
                  <Avatar id={conv.user_id || conv.conversation_id} name={name} url={conv.avatar_url || conv.group_avatar_url} />
                  <span className="conv-meta">
                    <span className="conv-name-row">
                      <span className="conv-name">{name}</span>
                      <time>{formatTime(conv.last_message_created_at || conv.updated_at)}</time>
                    </span>
                    <span className="conv-preview">{conv.last_message_content || 'No messages yet'}</span>
                  </span>
                </button>
              </li>
            )
          })
        )}
      </ul>

      <footer className="sidebar-foot">
        <Avatar id={user.id} name={displayName(user)} url={user.avatar_url} />
        <div className="conv-meta">
          <span className="conv-name">{displayName(user)}</span>
          <span className="conv-preview">@{user.username}</span>
        </div>
        <button type="button" className="logout-button" onClick={() => void onLogout()} aria-label="Log out" title="Log out">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="M10 5H6.5A1.5 1.5 0 0 0 5 6.5v11A1.5 1.5 0 0 0 6.5 19H10" />
            <path d="M14 8l4 4-4 4M9 12h9" />
          </svg>
        </button>
      </footer>
    </aside>
  )
}
