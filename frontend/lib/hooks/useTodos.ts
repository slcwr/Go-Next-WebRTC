import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { todosApi } from '@/lib/api/todos';
import type { CreateTodoRequest, UpdateTodoRequest } from '@/lib/types';

export function useTodos() {
  const queryClient = useQueryClient();

  // Todo一覧取得
  const { data: todos, isLoading, error } = useQuery({
    queryKey: ['todos'],
    queryFn: () => todosApi.getAll(),
    staleTime: 30 * 1000, // 30秒
  });

  // Todo作成
  const createMutation = useMutation({
    mutationFn: (data: CreateTodoRequest) => todosApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
    },
  });

  // Todo更新
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateTodoRequest }) =>
      todosApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
    },
  });

  // Todo削除
  const deleteMutation = useMutation({
    mutationFn: (id: number) => todosApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
    },
  });

  // Todo完了/未完了切り替え
  const toggleCompleteMutation = useMutation({
    mutationFn: ({ id, completed }: { id: number; completed: boolean }) =>
      todosApi.toggleComplete(id, completed),
    onMutate: async ({ id, completed }) => {
      // 楽観的更新
      await queryClient.cancelQueries({ queryKey: ['todos'] });
      const previousTodos = queryClient.getQueryData(['todos']);

      queryClient.setQueryData(['todos'], (old: any) =>
        old?.map((todo: any) =>
          todo.id === id ? { ...todo, completed } : todo
        )
      );

      return { previousTodos };
    },
    onError: (err, variables, context) => {
      // エラー時はロールバック
      if (context?.previousTodos) {
        queryClient.setQueryData(['todos'], context.previousTodos);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
    },
  });

  return {
    todos: todos || [],
    isLoading,
    error,
    createTodo: createMutation.mutateAsync,
    updateTodo: updateMutation.mutateAsync,
    deleteTodo: deleteMutation.mutateAsync,
    toggleComplete: toggleCompleteMutation.mutateAsync,
    isCreating: createMutation.isPending,
    isUpdating: updateMutation.isPending,
    isDeleting: deleteMutation.isPending,
    isToggling: toggleCompleteMutation.isPending,
  };
}
