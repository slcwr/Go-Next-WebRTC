'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/lib/hooks/useAuth';
import { useTodos } from '@/lib/hooks/useTodos';
import type { Todo } from '@/lib/types';

const todoSchema = z.object({
  title: z.string().min(1, 'タイトルを入力してください').max(255, 'タイトルは255文字以内で入力してください'),
  description: z.string().optional(),
});

type TodoFormData = z.infer<typeof todoSchema>;

function TodoList() {
  const { user, logout } = useAuth();
  const { todos, isLoading, createTodo, updateTodo, deleteTodo, toggleComplete } = useTodos();
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editForm, setEditForm] = useState<{ title: string; description: string }>({
    title: '',
    description: '',
  });

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<TodoFormData>({
    resolver: zodResolver(todoSchema),
  });

  const onSubmit = async (data: TodoFormData) => {
    try {
      await createTodo({
        title: data.title,
        description: data.description || '',
      });
      reset();
    } catch (error) {
      console.error('Failed to create todo:', error);
    }
  };

  const handleToggle = async (todo: Todo) => {
    try {
      await toggleComplete({ id: todo.id, completed: !todo.completed });
    } catch (error) {
      console.error('Failed to toggle todo:', error);
    }
  };

  const handleDelete = async (id: number) => {
    if (confirm('本当に削除しますか？')) {
      try {
        await deleteTodo(id);
      } catch (error) {
        console.error('Failed to delete todo:', error);
      }
    }
  };

  const startEdit = (todo: Todo) => {
    setEditingId(todo.id);
    setEditForm({ title: todo.title, description: todo.description });
  };

  const cancelEdit = () => {
    setEditingId(null);
    setEditForm({ title: '', description: '' });
  };

  const saveEdit = async (id: number) => {
    try {
      await updateTodo({ id, data: editForm });
      setEditingId(null);
    } catch (error) {
      console.error('Failed to update todo:', error);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <h1 className="text-xl font-bold text-gray-900">Todo List</h1>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-700">こんにちは、{user?.name}さん</span>
              <button
                onClick={() => logout()}
                className="text-sm text-gray-700 hover:text-gray-900"
              >
                ログアウト
              </button>
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          {/* 新規作成フォーム */}
          <div className="bg-white shadow rounded-lg p-6 mb-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">新しいTodoを作成</h2>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <div>
                <label htmlFor="title" className="block text-sm font-medium text-gray-700">
                  タイトル
                </label>
                <input
                  {...register('title')}
                  type="text"
                  id="title"
                  className="mt-1 block w-full rounded-md text-gray-700 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                  placeholder="タイトルを入力"
                />
                {errors.title && (
                  <p className="mt-1 text-sm text-red-600">{errors.title.message}</p>
                )}
              </div>
              <div>
                <label htmlFor="description" className="block text-sm font-medium text-gray-700">
                  説明（任意）
                </label>
                <textarea
                  {...register('description')}
                  id="description"
                  rows={3}
                  className="mt-1 block w-full rounded-md text-gray-700 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                  placeholder="説明を入力"
                />
              </div>
              <button
                type="submit"
                className="w-full sm:w-auto px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              >
                追加
              </button>
            </form>
          </div>

          {/* Todoリスト */}
          <div className="bg-white shadow rounded-lg">
            <div className="px-6 py-4 border-b border-gray-200">
              <h2 className="text-lg font-medium text-gray-900">Todoリスト</h2>
            </div>
            {isLoading ? (
              <div className="flex justify-center items-center py-12">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              </div>
            ) : todos.length === 0 ? (
              <div className="text-center py-12 text-gray-500">
                Todoがありません。新しいTodoを作成してください。
              </div>
            ) : (
              <ul className="divide-y divide-gray-200">
                {todos.map((todo) => (
                  <li key={todo.id} className="px-6 py-4 hover:bg-gray-50">
                    {editingId === todo.id ? (
                      <div className="space-y-3">
                        <input
                          type="text"
                          value={editForm.title}
                          onChange={(e) => setEditForm({ ...editForm, title: e.target.value })}
                          className="block w-full rounded-md text-gray-700 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                        />
                        <textarea
                          value={editForm.description}
                          onChange={(e) =>
                            setEditForm({ ...editForm, description: e.target.value })
                          }
                          rows={2}
                          className="block w-full rounded-md text-gray-700 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                        />
                        <div className="flex space-x-2">
                          <button
                            onClick={() => saveEdit(todo.id)}
                            className="px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
                          >
                            保存
                          </button>
                          <button
                            onClick={cancelEdit}
                            className="px-3 py-1 text-sm bg-gray-300 text-gray-700 rounded hover:bg-gray-400"
                          >
                            キャンセル
                          </button>
                        </div>
                      </div>
                    ) : (
                      <div className="flex items-start space-x-3">
                        <input
                          type="checkbox"
                          checked={todo.completed}
                          onChange={() => handleToggle(todo)}
                          className="mt-1 h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                        />
                        <div className="flex-1 min-w-0">
                          <p
                            className={`text-sm font-medium ${
                              todo.completed
                                ? 'line-through text-gray-500'
                                : 'text-gray-900'
                            }`}
                          >
                            {todo.title}
                          </p>
                          {todo.description && (
                            <p
                              className={`text-sm ${
                                todo.completed ? 'line-through text-gray-400' : 'text-gray-500'
                              }`}
                            >
                              {todo.description}
                            </p>
                          )}
                          <p className="text-xs text-gray-400 mt-1">
                            作成日: {new Date(todo.created_at).toLocaleDateString('ja-JP')}
                          </p>
                        </div>
                        <div className="flex space-x-2">
                          <button
                            onClick={() => startEdit(todo)}
                            className="text-sm text-blue-600 hover:text-blue-800"
                          >
                            編集
                          </button>
                          <button
                            onClick={() => handleDelete(todo.id)}
                            className="text-sm text-red-600 hover:text-red-800"
                          >
                            削除
                          </button>
                        </div>
                      </div>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}

export default function TodosPage() {
  return (
    <ProtectedRoute>
      <TodoList />
    </ProtectedRoute>
  );
}
