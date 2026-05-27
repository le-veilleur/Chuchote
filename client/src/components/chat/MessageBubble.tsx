import { useState, useRef, useEffect } from 'react';
import { Avatar } from '../ui/Avatar';
import type { Message } from '../../types/domain';
import { useAuthStore } from '../../store/auth.store';

interface Props {
  message: Message;
  onEdit: (messageId: string, content: string) => void;
  onDelete: (messageId: string) => void;
}

export function MessageBubble({ message, onEdit, onDelete }: Props) {
  const myId = useAuthStore((s) => s.userId);
  const isMine = message.authorId === myId;
  const [hovered, setHovered] = useState(false);
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(message.content);
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
      onMouseLeave={() => setHovered(false)}
    >
      {!isMine && <Avatar username={message.authorName} size={28} />}

      <div style={{ maxWidth: '70%', position: 'relative' }}>
        {/* Action buttons — only on own messages, on hover */}
        {isMine && hovered && !editing && !message.pending && (
          <div style={{
            position: 'absolute',
            top: -28,
            right: 0,
            display: 'flex',
            gap: 4,
            background: 'var(--color-bg)',
            border: '1px solid var(--color-border)',
            borderRadius: 8,
            padding: '2px 6px',
            zIndex: 10,
            whiteSpace: 'nowrap',
          }}>
            <button onClick={startEdit} style={actionBtnStyle} title="Modifier">✏️</button>
            <button onClick={() => onDelete(message.id)} style={{ ...actionBtnStyle, color: '#e74c3c' }} title="Supprimer">🗑️</button>
          </div>
        )}

        <div style={{
          background: isMine ? 'var(--color-accent)' : 'var(--color-bg-2)',
          color: isMine ? '#fff' : 'var(--color-text)',
          borderRadius: isMine ? '16px 16px 4px 16px' : '16px 16px 16px 4px',
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
                <button onClick={submitEdit} style={saveStyle}>Sauvegarder</button>
                <button onClick={cancelEdit} style={cancelStyle}>Annuler</button>
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
      </div>
    </div>
  );
}

const actionBtnStyle: React.CSSProperties = {
  background: 'none',
  border: 'none',
  cursor: 'pointer',
  fontSize: 13,
  padding: '0 2px',
  lineHeight: 1,
};

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
