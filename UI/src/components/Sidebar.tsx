import { useEffect, useMemo, useState } from 'react'
import { Avatar } from './Avatar'
import { ThemeToggle } from './ThemeToggle'
import {
  conversationTitle,
  formatTime,
  handle,
  personName,
} from '../lib/format'
import { api } from '../lib/api'
import type { Conversation, User, UserSearch } from '../lib/types'

type SidebarProps = {
  user: User
  conversations: Conversation[]
  activeId: string | null
  onSelect: (conversation: Conversation) => void
  onCreated: (conversationId: string, peer: UserSearch) => void
  onNewGroup: () => void
  onJoinGroup: () => void
  onOpenProfile: () => void
  onLogout: () => void
}

export function Sidebar({
  user,
  conversations,
  activeId,
  onSelect,
  onCreated,
  onNewGroup,
  onJoinGroup,
  onOpenProfile,
  onLogout,
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
    return conversations.filter((c) =>
      `${conversationTitle(c)} ${c.username}`.toLowerCase().includes(q),
    )
  }, [conversations, query, results.length])

  async function startChat(peer: UserSearch) {
    const res = await api.openDirect(peer.id)
    // `onCreated` opens the thread (and on mobile swaps the sidebar out for
    // it), so nothing here may reset that state afterwards.
    onCreated(res.conversation_id, peer)
    setQuery('')
    setResults([])
  }

  const myName = personName(user)
  const myHandle = handle(user.username, myName)

  return (
    <aside className="sidebar">
      <header className="sidebar-head">
        <div className="brand compact">
          <span className="brand-mark" aria-hidden>
            <img src="/logo.png" alt="" />
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

        <div className="sidebar-actions">
          <button type="button" className="side-action" onClick={onNewGroup}>
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <circle cx="9" cy="9" r="3.4" />
              <path d="M3.5 19c0-3 2.5-4.7 5.5-4.7s5.5 1.7 5.5 4.7" />
              <path d="M18 8v6M15 11h6" />
            </svg>
            New group
          </button>
          <button type="button" className="side-action" onClick={onJoinGroup}>
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M15 4h3.5A1.5 1.5 0 0 1 20 5.5v13a1.5 1.5 0 0 1-1.5 1.5H15" />
              <path d="M11 8l4 4-4 4M15 12H4" />
            </svg>
            Join with code
          </button>
        </div>
      </div>

      {results.length > 0 ? (
        <div className="result-block">
          <p className="list-label">People</p>
          <ul className="conv-list">
            {results.map((peer) => {
              const name = personName(peer)
              const peerHandle = handle(peer.username, name)

              return (
                <li key={peer.id}>
                  <button type="button" className="conv-item" onClick={() => void startChat(peer)}>
                    <Avatar id={peer.id} name={name} url={peer.avatar_url} />
                    <span className="conv-meta">
                      <span className="conv-name">{name}</span>
                      {peerHandle ? (
                        <span className="conv-preview">{peerHandle}</span>
                      ) : null}
                    </span>
                    <span className="new-chat-arrow" aria-hidden>→</span>
                  </button>
                </li>
              )
            })}
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
            const name = conversationTitle(conv)
            return (
              <li key={conv.conversation_id}>
                <button
                  type="button"
                  className={`conv-item ${activeId === conv.conversation_id ? 'active' : ''}`}
                  onClick={() => onSelect(conv)}
                >
                  <Avatar id={conv.user_id || conv.conversation_id} name={name} url={conv.avatar_url || conv.group_avatar_url} />
                  <span className="conv-meta">
                    <span className="conv-name-row">
                      <span className="conv-name">{name}</span>
                      <time>{formatTime(conv.last_message_created_at || conv.updated_at)}</time>
                    </span>
                    <span className="conv-preview">{conv.last_message_text || 'No messages yet'}</span>
                  </span>
                </button>
              </li>
            )
          })
        )}
      </ul>

      <footer className="sidebar-foot">
        <div className="profile-button">
          <button
            type="button"
            className="profile-avatar-button"
            onClick={onOpenProfile}
            aria-label="Edit profile"
            title="Edit profile"
          >
            <Avatar id={user.id} name={myName} url={user.avatar_url} />
          </button>
          <span className="conv-meta">
            <span className="conv-name">{myName}</span>
            {myHandle ? <span className="conv-preview">{myHandle}</span> : null}
          </span>
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
