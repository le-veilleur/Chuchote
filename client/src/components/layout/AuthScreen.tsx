import { useState } from 'react';
import { useAuth } from '../../hooks/useAuth';

export function AuthScreen() {
  const { login, register } = useAuth();
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      if (mode === 'login') await login(username, password);
      else await register(username, password);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erreur inconnue');
    }
  };

  return (
    <div style={{
      height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center',
      background: 'var(--color-bg)',
    }}>
      <form onSubmit={handleSubmit} style={{
        background: 'var(--color-bg-2)',
        padding: 32,
        borderRadius: 12,
        width: 320,
        display: 'flex',
        flexDirection: 'column',
        gap: 16,
        boxShadow: 'var(--color-shadow) 0 4px 24px',
      }}>
        <h2 style={{ margin: 0, textAlign: 'center' }}>
          {mode === 'login' ? 'Connexion' : 'Inscription'}
        </h2>

        <input
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          placeholder="Nom d'utilisateur"
          required
          style={inputStyle}
        />
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Mot de passe"
          required
          style={inputStyle}
        />

        {error && <div style={{ color: '#e55', fontSize: 13 }}>{error}</div>}

        <button type="submit" style={btnStyle}>
          {mode === 'login' ? 'Se connecter' : "S'inscrire"}
        </button>

        <button
          type="button"
          onClick={() => setMode(mode === 'login' ? 'register' : 'login')}
          style={{ border: 'none', background: 'none', cursor: 'pointer', fontSize: 13, color: 'var(--color-accent)' }}
        >
          {mode === 'login' ? "Pas encore de compte ? S'inscrire" : 'Déjà un compte ? Se connecter'}
        </button>
      </form>
    </div>
  );
}

const inputStyle: React.CSSProperties = {
  padding: '10px 12px', borderRadius: 8, border: '1px solid var(--color-border)',
  fontSize: 14, background: 'var(--color-bg)', color: 'var(--color-text)', outline: 'none',
};

const btnStyle: React.CSSProperties = {
  padding: '10px', borderRadius: 8, border: 'none', background: 'var(--color-accent)',
  color: '#fff', cursor: 'pointer', fontWeight: 600, fontSize: 15,
};
