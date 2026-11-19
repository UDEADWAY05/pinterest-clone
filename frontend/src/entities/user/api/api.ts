import { useQuery } from '@tanstack/react-query';
import { User, UserProfile } from '../model/types';
import { cookies } from 'next/headers';

const BACKEND_URL = process.env.NEXT_PUBLIC_BACKEND_URL ?? 'http://localhost:8080';
const baseUrl = typeof window === 'undefined' ? BACKEND_URL : '';

export const fetchCurrentUser = async (): Promise<User | null> => {
  const cookieStore = await cookies();
  const session = cookieStore.get('session')?.value;
  const headers: Record<string, string> = {};

  if (session) {
    headers['cookie'] = `session=${session}`;
  } else {
    return null;
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

export const fetchUserById = async (id: number): Promise<UserProfile | null> => {
  const cookieStore = await cookies();
  const sessionId = cookieStore.get('session')?.value;

  const headers: Record<string, string> = {};

  if (sessionId) {
    headers['cookie'] = `session=${sessionId}`;
  }
  console.log(headers);
  const res = await fetch(`${baseUrl}/api/users/${id}`, {
    credentials: 'include',
    headers,
    method: 'GET',
  });

  // 401 — пользователь не авторизован, это не ошибка
  if (res.status === 401) return null;
  if (!res.ok) throw new Error('Failed to fetch current user');

  const body: UserProfile = await res.json();
  console.log(body);
  return body;
};

export const useCurrentUser = (initialData?: User | null) => {
  return useQuery({
    queryKey: ['user', 'currentUser'],
    queryFn: () => fetchCurrentUser(),
    initialData: initialData ?? undefined,
  });
};
