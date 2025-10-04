/**
 * Call API Service
 * Backend APIとの通信を管理
 */

import { apiClient } from './client';

export interface CreateRoomRequest {
  name: string;
  maxParticipants?: number;
}

export interface CreateRoomResponse {
  room_id: string;
  name: string;
  invite_url: string;
}

export interface Room {
  id: string;
  name: string;
  createdBy: number;
  createdAt: string;
  maxParticipants: number;
  isActive: boolean;
  participantCount?: number;
}

export interface JoinRoomResponse {
  roomId: string;
  message: string;
}

export interface LeaveRoomResponse {
  message: string;
}

/**
 * ルームを作成
 */
export async function createRoom(data: CreateRoomRequest): Promise<CreateRoomResponse> {
  const response = await apiClient.post<CreateRoomResponse>('/api/calls/rooms', data);
  return response.data;
}

/**
 * アクティブなルーム一覧を取得
 */
export async function getActiveRooms(): Promise<Room[]> {
  const response = await apiClient.get<Room[]>('/api/calls/rooms');
  return response.data;
}

/**
 * ルーム情報を取得
 */
export async function getRoom(roomId: string): Promise<Room> {
  const response = await apiClient.get<Room>(`/api/calls/rooms/${roomId}`);
  return response.data;
}

/**
 * ルームに参加
 */
export async function joinRoom(roomId: string): Promise<JoinRoomResponse> {
  const response = await apiClient.post<JoinRoomResponse>(`/api/calls/rooms/${roomId}/join`);
  return response.data;
}

/**
 * ルームから退出
 */
export async function leaveRoom(roomId: string): Promise<LeaveRoomResponse> {
  const response = await apiClient.post<LeaveRoomResponse>(`/api/calls/rooms/${roomId}/leave`);
  return response.data;
}

/**
 * ルームを削除
 */
export async function deleteRoom(roomId: string): Promise<{ message: string }> {
  const response = await apiClient.delete<{ message: string }>(`/api/calls/rooms/${roomId}`);
  return response.data;
}
