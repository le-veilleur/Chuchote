interface Props {
  usernames: string[];
}

export function TypingIndicator({ usernames }: Props) {
  if (usernames.length === 0) return null;

  const label =
    usernames.length === 1
      ? `${usernames[0]} est en train d'écrire…`
      : `${usernames.join(', ')} écrivent…`;

  return (
    <div style={{ fontSize: 12, color: 'var(--color-text-muted)', padding: '4px 12px', fontStyle: 'italic' }}>
      {label}
    </div>
  );
}
