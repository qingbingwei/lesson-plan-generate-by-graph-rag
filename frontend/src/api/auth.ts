import api from './index';
import type { 
  LoginRequest, 
  RegisterRequest, 
  LoginResponse, 
  User, 
  ApiResponse 
} from '@/types';

/**
 * 用户登录
 */
export async function login(data: LoginRequest): Promise<LoginResponse> {
  const response = await api.post<ApiResponse<LoginResponse>>('/auth/login', data);
  return response.data.data;
}

/**
 * 用户注册
 */
export async function register(data: RegisterRequest): Promise<User> {
  const response = await api.post<ApiResponse<User>>('/auth/register', data);
  return response.data.data;
}

/**
 * 刷新 Token
 */
export async function refreshToken(refreshTokenValue: string): Promise<{ access_token: string; refresh_token: string }> {
  const response = await api.post<ApiResponse<{ access_token: string; refresh_token: string }>>(
    '/auth/refresh',
    { refresh_token: refreshTokenValue }
  );
  return response.data.data;
}

/**
 * 获取当前用户信息
 */
export async function getCurrentUser(): Promise<User> {
  const response = await api.get<ApiResponse<User>>('/users/profile');
  return response.data.data;
}

/**
 * 更新用户信息
 */
export async function updateProfile(data: Partial<User>): Promise<User> {
  const response = await api.put<ApiResponse<User>>('/users/profile', data);
  return response.data.data;
}

/**
 * 修改密码
 */
export async function changePassword(data: { oldPassword: string; newPassword: string }): Promise<void> {
  await api.post('/auth/change-password', data);
}

/**
 * 退出登录
 */
export async function logout(): Promise<void> {
  await api.post('/auth/logout');
}
