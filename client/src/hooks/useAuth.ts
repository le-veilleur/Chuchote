import { useEffect } from 'react';
import { useAuthStore } from '../store/auth.store';
import { authService } from '../services/auth.service';
import { wsService } from '../services/websocket.service';

const WS_URL = 'ws://localhost:8080/ws';

export function useAuth() {
  const { token, userId, username, setAuth, clearAuth } = useAuthStore();

  // Reconnect WS on page reload if the user is already authenticated
  useEffect(() => {
    if (token) wsService.connect(WS_URL);
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const login = async (uname: string, password: string) => {
    const res = await authService.login(uname, password);
    setAuth(res.token, res.userId, res.username);
    wsService.connect(WS_URL);
  };

  const register = async (uname: string, password: string) => {
    const res = await authService.register(uname, password);
    setAuth(res.token, res.userId, res.username);
    wsService.connect(WS_URL);
  };

  const logout = () => {
    wsService.disconnect();
    clearAuth();
  };

  return { token, userId, username, isAuthenticated: !!token, login, register, logout };
}
