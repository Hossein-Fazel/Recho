import { useEffect, useState } from 'react'
import { Avatar } from './Avatar'
import { InviteCode } from './InviteCode'
import { api } from '../lib/api'
import { UNNAMED_GROUP, displayName, handle, personName } from '../lib/format'
import type {
  Conversation,
  ConversationInfo,
  GroupMember,
} from '../lib/types'

type ConversationInfoPanelProps = {
  conversation: Conversation | null
  currentUserId: string
  open: boolean
  onClose: () => void
}

const roleLabel: Record<GroupMember['role'], string> = {
  owner: 'Owner',
  admin: 'Admin',
  member: '',
}

const EMPTY_PLACEHOLDER = 'Not set'

type InfoRowProps = {
  label: string
  value: string
}

function InfoRow({ label, value }: InfoRowProps) {
  const filled = value.trim().length > 0

  return (
    <div className="info-row">
      <dt className="info-label">{label}</dt>
      <dd className={`info-value${filled ? '' : ' info-value-empty'}`}>
        {filled ? value : EMPTY_PLACEHOLDER}
      </dd>
    </div>
  )
}

export function ConversationInfoPanel({
  conversation,
  currentUserId,
  open,
  onClose,
}: ConversationInfoPanelProps) {
  const [info, setInfo] = useState<ConversationInfo | null>(null)
  const [loading, setLoading] = useState(false)
  const [infoError, setInfoError] = useState('')
  const [members, setMembers] = useState<GroupMember[]>([])
  const [membersError, setMembersError] = useState('')
  const [membersLoading, setMembersLoading] = useState(false)
  const [nextCursor, setNextCursor] = useState('')

  const conversationId = conversation?.conversation_id ?? null
  const isGroup = conversation?.conversation_type === 'group'

  useEffect(() => {
    if (!open) return

    function onKey(event: KeyboardEvent) {
      if (event.key === 'Escape') onClose()
    }

    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [open, onClose])

  useEffect(() => {
    if (!open || !conversationId) return

    const targetId = conversationId
    let cancelled = false

    async function loadInfo() {
      setLoading(true)
      setInfoError('')
      setInfo(null)
      try {
        const data = await api.conversationInfo(targetId)
        if (!cancelled) setInfo(data)
      } catch {
        if (!cancelled) setInfoError('Could not load conversation info')
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    void loadInfo()

    return () => {
      cancelled = true
    }
  }, [open, conversationId])

  useEffect(() => {
    if (!open || !conversationId || !isGroup) return

    const targetId = conversationId
    let cancelled = false

    async function loadFirstPage() {
      setMembers([])
      setMembersError('')
      setNextCursor('')
      setMembersLoading(true)
      try {
        const result = await api.groupMembers(targetId)
        if (cancelled) return
        setMembers(result.members)
        setNextCursor(result.nextCursor)
      } catch {
        if (!cancelled) setMembersError('Could not load members')
      } finally {
        if (!cancelled) setMembersLoading(false)
      }
    }

    void loadFirstPage()

    return () => {
      cancelled = true
    }
  }, [open, conversationId, isGroup])

  async function loadMoreMembers() {
    if (!conversationId || !nextCursor || membersLoading) return

    setMembersLoading(true)
    try {
      const result = await api.groupMembers(conversationId, nextCursor)
      setMembers((current) => [...current, ...result.members])
      setNextCursor(result.nextCursor)
    } catch {
      setMembersError('Could not load more members')
    } finally {
      setMembersLoading(false)
    }
  }

  if (!open || !conversation) return null

  const directUser = info?.user ?? null
  const group = info?.group ?? null

  const username = (
    isGroup ? '' : directUser?.username || conversation.username || ''
  ).trim()

  const title = isGroup
    ? displayName({
        name: group?.name,
        group_name: conversation.group_name,
      }) || UNNAMED_GROUP
    : personName({
        display_name: directUser?.display_name || conversation.display_name,
        username,
      })

  const bio = ((isGroup ? group?.bio : directUser?.bio) ?? '').trim()

  const rawDisplayName = (
    directUser?.display_name ||
    conversation.display_name ||
    ''
  ).trim()

  const groupName = (group?.name || conversation.group_name || '').trim()

  // The backend only sends the invite code to members allowed to share it,
  // so its presence is the permission check.
  const inviteCode = (group?.invite_code || '').trim()

  // The avatar URL is only used to render the image in the hero, never shown
  // as a text row.
  const avatarUrl = isGroup
    ? group?.avatar_url || conversation.group_avatar_url
    : directUser?.avatar_url || conversation.avatar_url

  const avatarId = isGroup
    ? group?.id || conversation.conversation_id
    : directUser?.id || conversation.user_id || conversation.conversation_id

  return (
    <>
      <button
        type="button"
        className="info-backdrop"
        onClick={onClose}
        aria-label="Close conversation info"
        tabIndex={-1}
      />

      <aside
        className="info-panel"
        role="dialog"
        aria-modal="true"
        aria-label="Conversation info"
      >
        <header className="info-head">
          <button
            type="button"
            className="icon-button info-close"
            onClick={onClose}
            aria-label="Close conversation info"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M6 6l12 12M18 6 6 18" />
            </svg>
          </button>
          <span className="info-head-label">
            {isGroup ? 'Group info' : 'Profile'}
          </span>
        </header>

        <div className="info-scroll">
          <section className="info-hero">
            <Avatar
              id={avatarId}
              name={title}
              url={avatarUrl || undefined}
              size="xl"
            />
            <h2>{title}</h2>
            {isGroup ? (
              <p className="info-subtitle">
                Group
                {members.length > 0
                  ? ` · ${members.length} member${members.length === 1 ? '' : 's'}`
                  : ''}
              </p>
            ) : (
              <p className="info-subtitle">
                {handle(username) || 'Direct message'}
              </p>
            )}
          </section>

          {loading ? <p className="info-note">Loading…</p> : null}
          {infoError ? <p className="info-note error">{infoError}</p> : null}

          {!isGroup ? (
            <dl className="info-card">
              <InfoRow label="Name" value={rawDisplayName} />
              <InfoRow label="Username" value={handle(username)} />
              <InfoRow label="Bio" value={bio} />
            </dl>
          ) : (
            <>
              <dl className="info-card">
                <InfoRow label="Name" value={groupName} />
                <InfoRow label="About" value={bio} />
              </dl>

              {inviteCode ? <InviteCode code={inviteCode} /> : null}

              <section className="info-members">
                <p className="list-label">
                  Members{members.length > 0 ? ` — ${members.length}` : ''}
                </p>

                {membersError ? (
                  <p className="info-note error">{membersError}</p>
                ) : membersLoading && members.length === 0 ? (
                  <p className="info-note">Loading members…</p>
                ) : members.length === 0 ? (
                  <p className="info-note">No members yet.</p>
                ) : (
                  <ul className="info-member-list">
                    {members.map((member) => {
                      const role = roleLabel[member.role]
                      const isYou = member.user_id === currentUserId
                      const memberName = personName(member)

                      return (
                        <li key={member.user_id} className="info-member">
                          <Avatar
                            id={member.user_id}
                            name={memberName}
                            url={member.avatar_url}
                            size="md"
                          />
                          <div className="info-member-meta">
                            <span className="info-member-name">
                              {memberName}
                              {isYou ? <span className="info-you">you</span> : null}
                              {role ? (
                                <span
                                  className={`role-badge role-${member.role}`}
                                >
                                  {role}
                                </span>
                              ) : null}
                            </span>
                            <span className="info-member-handle">
                              {member.username.trim()
                                ? `@${member.username.trim()}`
                                : EMPTY_PLACEHOLDER}
                            </span>
                          </div>
                        </li>
                      )
                    })}
                  </ul>
                )}

                {nextCursor ? (
                  <button
                    type="button"
                    className="ghost-btn info-more"
                    onClick={() => void loadMoreMembers()}
                    disabled={membersLoading}
                  >
                    {membersLoading ? 'Loading…' : 'Show more'}
                  </button>
                ) : null}
              </section>
            </>
          )}
        </div>
      </aside>
    </>
  )
}
