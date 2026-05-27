import type { ReactNode } from 'react';
import { Sidebar } from './Sidebar';

interface Props {
  children: ReactNode;
}

export function AppShell({ children }: Props) {
  return (
    <div style={{
      display: 'flex',
      height: '100vh',
      overflow: 'hidden',
      background: 'var(--color-bg)',
      color: 'var(--color-text)',
    }}>
      <Sidebar />
      {children}
    </div>
  );
}
