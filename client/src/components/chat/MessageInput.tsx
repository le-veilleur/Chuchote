import { useRef, useState } from 'react';
import { useMessages } from '../../hooks/useMessages';
import { useTypingIndicator } from '../../hooks/useTypingIndicator';

interface Props {
  roomId: string;
}

export function MessageInput({ roomId }: Props) {
  const [value, setValue] = useState('');
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
        disabled={!value.trim()}
        style={{
          padding: '8px 16px',
          borderRadius: 8,
          border: 'none',
          background: 'var(--color-accent)',
          color: '#fff',
          cursor: 'pointer',
          fontWeight: 600,
          fontSize: 14,
          opacity: value.trim() ? 1 : 0.4,
        }}
      >
        Envoyer
      </button>
    </div>
  );
}
