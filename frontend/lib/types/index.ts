// User types
export interface User {
  id: number;
  email: string;
  name: string;
  avatar_url?: string;
  bio?: string;
  is_active: boolean;
  email_verified_at?: string;
  created_at: string;
  updated_at: string;
}

// Auth types
export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  user: User;
  expires_at: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
}

export interface RefreshTokenRequest {
  refreshToken: string;
}

// Todo types
export interface Todo {
  id: number;
  title: string;
  description: string;
  completed: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateTodoRequest {
  title: string;
  description: string;
}

export interface UpdateTodoRequest {
  title?: string;
  description?: string;
  completed?: boolean;
}

// API Error type
export interface ApiError {
  message: string;
  code?: string;
  details?: Record<string, string[]>;
}
