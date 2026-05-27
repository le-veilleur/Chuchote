import { useRef, useState } from 'react';
import { SendHorizonal } from 'lucide-react';
import { useMessages } from '../../hooks/useMessages';
import { useTypingIndicator } from '../../hooks/useTypingIndicator';

interface Props {
  roomId: string;
}

export function MessageInput({ roomId }: Props) {
  const [value, setValue] = useState('');
  const [hovered, setHovered] = useState(false);
  const { send } = useMessages(roomId);
  const { onKeystroke } = useTypingIndicator(roomId);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleSubmit = () => {
    const trimmed = value.trim();
    if (!trimmed) return;
    send(trimmed);
    setValue('');
    textareaRef.current?.focus();
  };

  const active = !!value.trim();

  return (
    <div style={{
      display: 'flex',
      gap: 8,
      padding: '12px 16px',
      borderTop: '1px solid var(--color-border)',
      background: 'var(--color-bg)',
    }}>
      <textarea
        ref={textareaRef}
        value={value}
        onChange={(e) => { setValue(e.target.value); onKeystroke(); }}
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
          borderRadius: 8,
          border: '1px solid var(--color-border)',
          padding: '8px 12px',
          fontSize: 14,
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
  );
}
