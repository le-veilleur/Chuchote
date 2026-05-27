import { http } from './http.service';
import type { TokenResponse } from '../types/api';

export const authService = {
  register: (username: string, password: string) =>
    http.post<TokenResponse>('/auth/register', { username, password }),

  login: (username: string, password: string) =>
    http.post<TokenResponse>('/auth/login', { username, password }),
};
