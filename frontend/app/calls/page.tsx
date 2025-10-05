'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { getActiveRooms, createRoom, joinRoom, deleteRoom, type Room } from '@/lib/api/calls';
import Navigation from '@/components/Navigation';
import { useAuth } from '@/lib/hooks/useAuth';

export default function CallsPage() {
  const router = useRouter();
  const { user } = useAuth();
  const [rooms, setRooms] = useState<Room[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newRoomName, setNewRoomName] = useState('');
  const [creating, setCreating] = useState(false);
  const [deleting, setDeleting] = useState<string | null>(null);

  useEffect(() => {
    loadRooms();
  }, []);

  /**
   * ルーム一覧を読み込み
   */
  const loadRooms = async () => {
    try {
      setLoading(true);
      setError(null);

      const activeRooms = await getActiveRooms();
      setRooms(activeRooms);
    } catch (err) {
      console.error('Failed to load rooms:', err);
      setError(err instanceof Error ? err.message : 'Failed to load rooms');
      // 認証エラーの場合はログインページへ（apiClientのインターセプターが処理）
    } finally {
      setLoading(false);
    }
  };

  /**
   * 新しいルームを作成
   */
  const handleCreateRoom = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!newRoomName.trim()) {
      return;
    }

    try {
      setCreating(true);
      setError(null);

      const room = await createRoom({
        name: newRoomName,
        maxParticipants: 10,
      });

      // ルーム作成後、すぐに参加
      await joinRoom(room.room_id);
      router.push(`/calls/${room.room_id}`);
    } catch (err) {
      console.error('Failed to create room:', err);
      setError(err instanceof Error ? err.message : 'Failed to create room');
    } finally {
      setCreating(false);
    }
  };

  /**
   * ルームに参加
   */
  const handleJoinRoom = async (roomId: string) => {
    try {
      await joinRoom(roomId);
      // バックエンドでの参加処理完了を待つ
      await new Promise(resolve => setTimeout(resolve, 300));
      router.push(`/calls/${roomId}`);
    } catch (err) {
      console.error('Failed to join room:', err);
      setError(err instanceof Error ? err.message : 'Failed to join room');
    }
  };

  /**
   * ルームを削除
   */
  const handleDeleteRoom = async (roomId: string) => {
    if (!confirm('Are you sure you want to delete this room?')) {
      return;
    }

    try {
      setDeleting(roomId);
      await deleteRoom(roomId);
      // ルーム一覧を再読み込み
      await loadRooms();
    } catch (err) {
      console.error('Failed to delete room:', err);
      setError(err instanceof Error ? err.message : 'Failed to delete room');
    } finally {
      setDeleting(null);
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      {/* ナビゲーション */}
      <Navigation />

      {/* ヘッダー */}
      <div className="bg-gray-800 border-b border-gray-700 p-4">
        <div className="container mx-auto flex justify-between items-center">
          <h1 className="text-2xl font-bold">Video Calls</h1>
          <button
            onClick={() => setShowCreateModal(true)}
            className="bg-blue-500 hover:bg-blue-600 px-4 py-2 rounded font-medium transition"
          >
            + New Room
          </button>
        </div>
      </div>

      {/* メインコンテンツ */}
      <div className="container mx-auto p-4">
        {error && (
          <div className="bg-red-500 text-white p-4 rounded mb-4">
            {error}
            <button
              onClick={() => setError(null)}
              className="ml-4 underline"
            >
              Dismiss
            </button>
          </div>
        )}

        {loading ? (
          <div className="flex justify-center items-center h-64">
            <div className="text-gray-400">Loading rooms...</div>
          </div>
        ) : rooms.length === 0 ? (
          <div className="text-center py-12">
            <div className="text-6xl mb-4">📹</div>
            <h2 className="text-xl font-semibold mb-2">No active rooms</h2>
            <p className="text-gray-400 mb-6">Create a new room to get started</p>
            <button
              onClick={() => setShowCreateModal(true)}
              className="bg-blue-500 hover:bg-blue-600 px-6 py-3 rounded font-medium transition"
            >
              Create Room
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {rooms.map((room) => (
              <div
                key={room.id}
                className="bg-gray-800 rounded-lg p-6 border border-gray-700 hover:border-gray-600 transition"
              >
                <h3 className="text-lg font-semibold mb-2">{room.name}</h3>
                <div className="text-sm text-gray-400 mb-4">
                  <div>Participants: {room.participantCount || 0} / {room.maxParticipants}</div>
                  <div>Created: {new Date(room.createdAt).toLocaleDateString()}</div>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => handleJoinRoom(room.id)}
                    className="flex-1 bg-blue-500 hover:bg-blue-600 px-4 py-2 rounded font-medium transition"
                  >
                    Join Room
                  </button>
                  {user?.id === room.createdBy && (
                    <button
                      onClick={() => handleDeleteRoom(room.id)}
                      disabled={deleting === room.id}
                      className="bg-white hover:bg-gray-200 disabled:bg-gray-600 disabled:cursor-not-allowed px-4 py-2 rounded font-medium transition text-gray-900"
                      title="Delete room"
                    >
                      {deleting === room.id ? '...' : '🗑️'}
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* ルーム作成モーダル */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-gray-800 rounded-lg p-6 max-w-md w-full">
            <h2 className="text-xl font-bold mb-4">Create New Room</h2>
            <form onSubmit={handleCreateRoom}>
              <div className="mb-4">
                <label className="block text-sm font-medium mb-2">
                  Room Name
                </label>
                <input
                  type="text"
                  value={newRoomName}
                  onChange={(e) => setNewRoomName(e.target.value)}
                  className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded focus:outline-none focus:border-blue-500"
                  placeholder="Enter room name"
                  autoFocus
                  required
                />
              </div>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => {
                    setShowCreateModal(false);
                    setNewRoomName('');
                  }}
                  className="flex-1 px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded font-medium transition"
                  disabled={creating}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-2 bg-blue-500 hover:bg-blue-600 rounded font-medium transition disabled:opacity-50 disabled:cursor-not-allowed"
                  disabled={creating}
                >
                  {creating ? 'Creating...' : 'Create & Join'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
