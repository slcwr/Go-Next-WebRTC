import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { authApi } from '@/lib/api/auth';
import { useAuthStore } from '@/lib/stores/authStore';
import type { LoginRequest, RegisterRequest } from '@/lib/types';

export function useAuth() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { user, isAuthenticated, setAuth, clearAuth } = useAuthStore();

  // ログイン
  const loginMutation = useMutation({
    mutationFn: (data: LoginRequest) => authApi.login(data),
    onSuccess: (data) => {
      setAuth(data.user, data.access_token, data.refresh_token);
      // ユーザー情報をキャッシュに設定
      queryClient.setQueryData(['user'], data.user);
      // 状態が確実に更新された後にリダイレクト
      setTimeout(() => router.push('/todos'), 100);
    },
  });

  // 新規登録
  const registerMutation = useMutation({
    mutationFn: (data: RegisterRequest) => authApi.register(data),
    onSuccess: (data) => {
      setAuth(data.user, data.access_token, data.refresh_token);
      // ユーザー情報をキャッシュに設定
      queryClient.setQueryData(['user'], data.user);
      // 状態が確実に更新された後にリダイレクト
      setTimeout(() => router.push('/todos'), 100);
    },
  });

  // ログアウト
  const logoutMutation = useMutation({
    mutationFn: () => authApi.logout(),
    onSuccess: () => {
      clearAuth();
      queryClient.clear();
      router.push('/login');
    },
    onError: () => {
      // エラーが発生してもローカルの状態はクリア
      clearAuth();
      queryClient.clear();
      router.push('/login');
    },
  });

  // ユーザー情報取得（ページリロード時のみ）
  const { data: currentUser, isLoading: isUserLoading } = useQuery({
    queryKey: ['user'],
    queryFn: () => authApi.getMe(),
    enabled: isAuthenticated && typeof window !== 'undefined' && !!localStorage.getItem('accessToken'),
    retry: false,
    staleTime: 5 * 60 * 1000, // 5分
  });

  return {
    user: currentUser || user,
    isAuthenticated,
    isLoading: isUserLoading,
    login: loginMutation.mutateAsync,
    register: registerMutation.mutateAsync,
    logout: logoutMutation.mutate,
    isLoginLoading: loginMutation.isPending,
    isRegisterLoading: registerMutation.isPending,
    isLogoutLoading: logoutMutation.isPending,
    loginError: loginMutation.error,
    registerError: registerMutation.error,
  };
}
