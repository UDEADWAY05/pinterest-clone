import { useQuery } from '@tanstack/react-query';
import { User } from '../model/types';

const BACKEND_URL = process.env.NEXT_PUBLIC_BACKEND_URL ?? 'http://localhost:8080';
const baseUrl = typeof window === 'undefined' ? BACKEND_URL : '';

export const fetchCurrentUser = async (cookieHeader?: string): Promise<User | null> => {
  const headers: Record<string, string> = {};

  if (cookieHeader) {
    headers['cookie'] = cookieHeader;
  }

  const res = await fetch(`${baseUrl}/api/user/current`, {
    credentials: 'include',
    headers,
    method: 'GET',
  });

  // 401 — пользователь не авторизован, это не ошибка
  if (res.status === 401) return null;
  if (!res.ok) throw new Error('Failed to fetch current user');

  const body: User = await res.json();
  console.log(body);
  return body;
};

export const useCurrentUser = (initialData?: User | null, cookieHeader?: string) => {
  return useQuery({
    queryKey: ['user', 'currentUser'],
    queryFn: () => fetchCurrentUser(cookieHeader),
    initialData: initialData ?? undefined,
  });
};
