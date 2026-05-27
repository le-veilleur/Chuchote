import { useState, useRef, useEffect } from 'react';
import { Pencil, Trash2, Check, X, Reply } from 'lucide-react';
import { Avatar } from '../ui/Avatar';
import type { Message } from '../../types/domain';
import { useAuthStore } from '../../store/auth.store';

const REACTION_EMOJIS = ['👍', '❤️', '😂', '😮', '😢', '🎉'];

interface Props {
  message: Message;
  onEdit: (messageId: string, content: string) => void;
  onDelete: (messageId: string) => void;
  onReply: (message: Message) => void;
  onReaction: (messageId: string, emoji: string) => void;
}

export function MessageBubble({ message, onEdit, onDelete, onReply, onReaction }: Props) {
  const myId = useAuthStore((s) => s.userId);
  const isMine = message.authorId === myId;
  const [hovered, setHovered] = useState(false);
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(message.content);
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (editing) {
      textareaRef.current?.focus();
      textareaRef.current?.select();
    }
  }, [editing]);

  const startEdit = () => {
    setDraft(message.content);
    setEditing(true);
  };

  const submitEdit = () => {
    if (draft.trim() && draft !== message.content) {
      onEdit(message.id, draft.trim());
    }
    setEditing(false);
  };

  const cancelEdit = () => {
    setDraft(message.content);
    setEditing(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); submitEdit(); }
    if (e.key === 'Escape') cancelEdit();
  };

  const reactions = message.reactions ?? [];

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: isMine ? 'row-reverse' : 'row',
        gap: 8,
        alignItems: 'flex-end',
        opacity: message.pending ? 0.6 : 1,
        position: 'relative',
      }}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => { setHovered(false); setShowEmojiPicker(false); }}
    >
      {!isMine && <Avatar username={message.authorName} size={28} />}

      <div style={{ maxWidth: '70%', position: 'relative' }}>
        {/* Action buttons on hover */}
        {hovered && !editing && !message.pending && (
          <div style={{
            position: 'absolute',
            top: -32,
            right: isMine ? 0 : 'auto',
            left: isMine ? 'auto' : 0,
            display: 'flex',
            gap: 2,
            background: 'var(--color-bg)',
            border: '1px solid var(--color-border)',
            borderRadius: 8,
            padding: '2px 4px',
            zIndex: 10,
            whiteSpace: 'nowrap',
          }}>
            {/* Emoji picker trigger */}
            <div style={{ position: 'relative' }}>
              <button
                onClick={() => setShowEmojiPicker((v) => !v)}
                style={emojiPickerBtnStyle(showEmojiPicker)}
                title="Réagir"
              >
                😊
              </button>
              {showEmojiPicker && (
                <div style={{
                  position: 'absolute',
                  top: -40,
                  left: 0,
                  display: 'flex',
                  gap: 2,
                  background: 'var(--color-bg)',
                  border: '1px solid var(--color-border)',
                  borderRadius: 8,
                  padding: '4px 6px',
                  zIndex: 20,
                  boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
                }}>
                  {REACTION_EMOJIS.map((emoji) => (
                    <button
                      key={emoji}
                      onClick={() => { onReaction(message.id, emoji); setShowEmojiPicker(false); }}
                      style={{
                        background: 'none', border: 'none', cursor: 'pointer',
                        fontSize: 18, padding: '0 3px', lineHeight: 1,
                        borderRadius: 4,
                        transition: 'transform 0.1s',
                      }}
                      onMouseEnter={(e) => (e.currentTarget.style.transform = 'scale(1.25)')}
                      onMouseLeave={(e) => (e.currentTarget.style.transform = 'scale(1)')}
                    >
                      {emoji}
                    </button>
                  ))}
                </div>
              )}
            </div>

            <ActionIconButton onClick={() => onReply(message)} title="Répondre">
              <Reply size={13} />
            </ActionIconButton>

            {isMine && (
              <>
                <ActionIconButton onClick={startEdit} title="Modifier">
                  <Pencil size={13} />
                </ActionIconButton>
                <ActionIconButton onClick={() => onDelete(message.id)} title="Supprimer" danger>
                  <Trash2 size={13} />
                </ActionIconButton>
              </>
            )}
          </div>
        )}

        {/* Reply preview */}
        {message.replyToSummary && (
          <div style={{
            background: 'rgba(255,255,255,0.06)',
            borderLeft: '3px solid var(--color-accent)',
            borderRadius: '6px 6px 0 0',
            padding: '5px 10px',
            fontSize: 12,
            marginBottom: -2,
            opacity: 0.85,
            maxWidth: '100%',
            overflow: 'hidden',
          }}>
            <span style={{ fontWeight: 600, color: 'var(--color-accent)' }}>
              {message.replyToSummary.authorName}
            </span>
            <span style={{ marginLeft: 6, opacity: 0.8, display: 'block', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
              {message.replyToSummary.content}
            </span>
          </div>
        )}

        <div style={{
          background: isMine ? 'var(--color-accent)' : 'var(--color-bg-2)',
          color: isMine ? '#fff' : 'var(--color-text)',
          borderRadius: message.replyToSummary
            ? (isMine ? '16px 4px 4px 16px' : '4px 16px 16px 4px')
            : (isMine ? '16px 16px 4px 16px' : '16px 16px 16px 4px'),
          padding: '8px 12px',
          fontSize: 14,
          lineHeight: 1.5,
          wordBreak: 'break-word',
        }}>
          {!isMine && (
            <div style={{ fontSize: 11, fontWeight: 600, marginBottom: 2, opacity: 0.7 }}>
              {message.authorName}
            </div>
          )}

          {editing ? (
            <div>
              <textarea
                ref={textareaRef}
                value={draft}
                onChange={(e) => setDraft(e.target.value)}
                onKeyDown={handleKeyDown}
                rows={Math.min(draft.split('\n').length + 1, 6)}
                style={{
                  width: '100%',
                  background: 'rgba(255,255,255,0.15)',
                  color: 'inherit',
                  border: '1px solid rgba(255,255,255,0.4)',
                  borderRadius: 6,
                  padding: '4px 6px',
                  fontSize: 14,
                  resize: 'none',
                  outline: 'none',
                  fontFamily: 'inherit',
                }}
              />
              <div style={{ display: 'flex', gap: 6, marginTop: 4, fontSize: 11 }}>
                <button onClick={submitEdit} style={saveStyle}>
                  <Check size={11} style={{ display: 'inline', verticalAlign: 'middle', marginRight: 3 }} />
                  Sauvegarder
                </button>
                <button onClick={cancelEdit} style={cancelStyle}>
                  <X size={11} style={{ display: 'inline', verticalAlign: 'middle', marginRight: 3 }} />
                  Annuler
                </button>
              </div>
              <div style={{ fontSize: 10, opacity: 0.6, marginTop: 2 }}>Entrée pour sauvegarder · Échap pour annuler</div>
            </div>
          ) : (
            <div>{message.content}</div>
          )}

          <div style={{ fontSize: 10, marginTop: 4, opacity: 0.6, textAlign: 'right' }}>
            {new Date(message.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            {message.editedAt && <span style={{ marginLeft: 4 }}>(modifié)</span>}
          </div>
        </div>

        {/* Reaction pills */}
        {reactions.length > 0 && (
          <div style={{
            display: 'flex',
            flexWrap: 'wrap',
            gap: 4,
            marginTop: 4,
            justifyContent: isMine ? 'flex-end' : 'flex-start',
          }}>
            {reactions.map((r) => {
              const iReacted = r.userIds.includes(myId ?? '');
              return (
                <button
                  key={r.emoji}
                  onClick={() => onReaction(message.id, r.emoji)}
                  style={{
                    display: 'flex', alignItems: 'center', gap: 3,
                    background: iReacted ? 'rgba(139,92,246,0.25)' : 'var(--color-bg-2)',
                    border: iReacted ? '1px solid var(--color-accent)' : '1px solid var(--color-border)',
                    borderRadius: 12,
                    padding: '2px 7px',
                    cursor: 'pointer',
                    fontSize: 13,
                    lineHeight: 1.4,
                    transition: 'background 0.1s',
                  }}
                >
                  <span>{r.emoji}</span>
                  <span style={{ fontSize: 11, fontWeight: 600, color: 'var(--color-text-muted)' }}>{r.count}</span>
                </button>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}

function ActionIconButton({ onClick, title, danger, children }: {
  onClick: () => void;
  title: string;
  danger?: boolean;
  children: React.ReactNode;
}) {
  const [hovered, setHovered] = useState(false);
  return (
    <button
      onClick={onClick}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      title={title}
      style={{
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        width: 24, height: 24,
        background: hovered ? (danger ? 'rgba(231,76,60,0.15)' : 'rgba(255,255,255,0.1)') : 'none',
        border: 'none',
        borderRadius: 4,
        cursor: 'pointer',
        color: hovered && danger ? '#e74c3c' : 'var(--color-text-muted)',
        transition: 'background 0.1s, color 0.1s',
        padding: 0,
      }}
    >
      {children}
    </button>
  );
}

function emojiPickerBtnStyle(active: boolean): React.CSSProperties {
  return {
    display: 'flex', alignItems: 'center', justifyContent: 'center',
    width: 24, height: 24,
    background: active ? 'rgba(255,255,255,0.1)' : 'none',
    border: 'none',
    borderRadius: 4,
    cursor: 'pointer',
    fontSize: 14,
    padding: 0,
    transition: 'background 0.1s',
  };
}

const saveStyle: React.CSSProperties = {
  background: 'rgba(255,255,255,0.25)',
  border: 'none',
  borderRadius: 4,
  color: 'inherit',
  cursor: 'pointer',
  padding: '2px 8px',
  fontSize: 11,
};

const cancelStyle: React.CSSProperties = {
  background: 'none',
  border: 'none',
  color: 'inherit',
  cursor: 'pointer',
  opacity: 0.7,
  padding: '2px 4px',
  fontSize: 11,
};
