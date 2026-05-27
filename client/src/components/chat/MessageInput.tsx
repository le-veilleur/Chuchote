import { useRef, useState } from 'react';
import { SendHorizonal, X } from 'lucide-react';
import { useMessages } from '../../hooks/useMessages';
import { useTypingIndicator } from '../../hooks/useTypingIndicator';
import type { Message } from '../../types/domain';

interface Props {
  roomId: string;
  replyingTo: Message | null;
  onCancelReply: () => void;
}

export function MessageInput({ roomId, replyingTo, onCancelReply }: Props) {
  const [value, setValue] = useState('');
  const [hovered, setHovered] = useState(false);
  const { send } = useMessages(roomId);
  const { onKeystroke } = useTypingIndicator(roomId);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleSubmit = () => {
    const trimmed = value.trim();
    if (!trimmed) return;
    send(trimmed, replyingTo?.id);
    setValue('');
    onCancelReply();
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.focus();
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setValue(e.target.value);
    onKeystroke();
    e.target.style.height = 'auto';
    e.target.style.height = `${e.target.scrollHeight}px`;
  };

  const active = !!value.trim();

  return (
    <div style={{
      borderTop: '1px solid var(--color-border)',
      background: 'var(--color-bg)',
    }}>
      {replyingTo && (
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          padding: '8px 16px 0',
          fontSize: 12,
          color: 'var(--color-text-muted)',
          borderLeft: '3px solid var(--color-accent)',
          marginLeft: 16,
          marginRight: 16,
          marginTop: 8,
        }}>
          <div style={{ flex: 1, minWidth: 0 }}>
            <span style={{ fontWeight: 600, color: 'var(--color-accent)' }}>
              Répondre à {replyingTo.authorName}
            </span>
            <span style={{ marginLeft: 6, opacity: 0.7, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', display: 'inline-block', maxWidth: 300, verticalAlign: 'bottom' }}>
              {replyingTo.content}
            </span>
          </div>
          <button
            onClick={onCancelReply}
            style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--color-text-muted)', padding: 2, display: 'flex' }}
          >
            <X size={14} />
          </button>
        </div>
      )}
      <div style={{
        display: 'flex',
        gap: 8,
        padding: '12px 16px',
        alignItems: 'flex-end',
      }}>
        <textarea
          ref={textareaRef}
          value={value}
          onChange={handleChange}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault();
              handleSubmit();
            }
          }}
          placeholder="Écrire un message… (Entrée pour envoyer)"
          rows={1}
          style={{
            flex: 1,
            resize: 'none',
            overflow: 'hidden',
            borderRadius: 8,
            border: '1px solid var(--color-border)',
            padding: '8px 12px',
            fontSize: 14,
            lineHeight: '1.5',
            fontFamily: 'inherit',
            background: 'var(--color-bg-2)',
            color: 'var(--color-text)',
            outline: 'none',
          }}
        />
        <button
          onClick={handleSubmit}
          onMouseEnter={() => setHovered(true)}
          onMouseLeave={() => setHovered(false)}
          disabled={!active}
          title="Envoyer"
          style={{
            display: 'flex', alignItems: 'center', gap: 6,
            padding: '8px 16px',
            borderRadius: 8,
            border: 'none',
            background: active && hovered ? 'var(--color-accent-hover, #7c3aed)' : 'var(--color-accent)',
            color: '#fff',
            cursor: active ? 'pointer' : 'not-allowed',
            fontWeight: 600,
            fontSize: 14,
            opacity: active ? 1 : 0.4,
            transition: 'background 0.15s, opacity 0.15s',
          }}
        >
          <SendHorizonal size={16} />
          Envoyer
        </button>
      </div>
    </div>
  );
}
