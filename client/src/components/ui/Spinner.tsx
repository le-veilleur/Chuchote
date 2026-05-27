export function Spinner() {
  return (
    <div style={{
      width: 20, height: 20, borderRadius: '50%',
      border: '2px solid var(--color-border)',
      borderTopColor: 'var(--color-accent)',
      animation: 'spin 0.7s linear infinite',
    }} />
  );
}
