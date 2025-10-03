import { apiClient } from './client';
import type { Todo, CreateTodoRequest, UpdateTodoRequest } from '@/lib/types';

export const todosApi = {
  // Todo一覧取得
  getAll: async (): Promise<Todo[]> => {
    const response = await apiClient.get<Todo[]>('/api/todos');
    return response.data;
  },

  // Todo詳細取得
  getById: async (id: number): Promise<Todo> => {
    const response = await apiClient.get<Todo>(`/api/todos/${id}`);
    return response.data;
  },

  // Todo作成
  create: async (data: CreateTodoRequest): Promise<Todo> => {
    const response = await apiClient.post<Todo>('/api/todos', data);
    return response.data;
  },

  // Todo更新
  update: async (id: number, data: UpdateTodoRequest): Promise<Todo> => {
    console.log('Updating todo:', { id, data });
    const response = await apiClient.put<Todo>(`/api/todos/${id}`, data);
    return response.data;
  },

  // Todo削除
  delete: async (id: number): Promise<void> => {
    await apiClient.delete(`/api/todos/${id}`);
  },

  // Todo完了/未完了切り替え
  toggleComplete: async (id: number, completed: boolean): Promise<Todo> => {
    const response = await apiClient.put<Todo>(`/api/todos/${id}`, { completed });
    return response.data;
  },
};
