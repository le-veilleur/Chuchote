import { useAuthStore } from '../store/auth.store';
import { authService } from '../services/auth.service';
import { wsService } from '../services/websocket.service';

export function useAuth() {
  const { token, userId, username, setAuth, clearAuth } = useAuthStore();

  const login = async (uname: string, password: string) => {
    const res = await authService.login(uname, password);
    setAuth(res.token, res.userId, res.username);
    wsService.connect('ws://localhost:8080/ws');
    wsService.send({ type: 'auth.connect', requestId: crypto.randomUUID(), roomId: null, payload: { token: res.token } });
  };

  const register = async (uname: string, password: string) => {
    const res = await authService.register(uname, password);
    setAuth(res.token, res.userId, res.username);
    wsService.connect('ws://localhost:8080/ws');
    wsService.send({ type: 'auth.connect', requestId: crypto.randomUUID(), roomId: null, payload: { token: res.token } });
  };

  const logout = () => {
    wsService.disconnect();
    clearAuth();
  };

  return { token, userId, username, isAuthenticated: !!token, login, register, logout };
}
