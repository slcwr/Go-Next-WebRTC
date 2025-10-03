import { apiClient } from './client';
import type {
  AuthTokens,
  LoginRequest,
  RegisterRequest,
  User,
} from '@/lib/types';

export const authApi = {
  // 新規登録
  register: async (data: RegisterRequest): Promise<AuthTokens> => {
    const response = await apiClient.post<AuthTokens>('/api/auth/register', data);
    return response.data;
  },

  // ログイン
  login: async (data: LoginRequest): Promise<AuthTokens> => {
    const response = await apiClient.post<AuthTokens>('/api/auth/login', data);
    return response.data;
  },

  // ログアウト
  logout: async (): Promise<void> => {
    await apiClient.post('/api/auth/logout');
  },

  // トークンリフレッシュ
  refreshToken: async (refreshToken: string): Promise<AuthTokens> => {
    const response = await apiClient.post<AuthTokens>('/api/auth/refresh', {
      refresh_token: refreshToken,
    });
    return response.data;
  },

  // ユーザー情報取得
  getMe: async (): Promise<User> => {
    const response = await apiClient.get<User>('/api/auth/me');
    return response.data;
  },

  // パスワード変更
  changePassword: async (oldPassword: string, newPassword: string): Promise<void> => {
    await apiClient.post('/api/auth/password', {
      old_password: oldPassword,
      new_password: newPassword,
    });
  },
};
