import { useQuery } from '@tanstack/react-query';
import { User } from '../model/types';

const BACKEND_URL = process.env.NEXT_PUBLIC_BACKEND_URL ?? 'http://localhost:8080';

export const fetchCurrentUser = async (cookieHeader?: string): Promise<User | null> => {
  const headers: Record<string, string> = {};

  // Если вызов на сервере — прокидываем переданный заголовок
  if (cookieHeader) {
    headers['cookie'] = cookieHeader;
  }

  const res = await fetch(`${typeof window === 'undefined' ? BACKEND_URL : ''}/api/user/current`, {
    // cache: 'no-store',
    // credentials важен для клиента (если запросы идут на тот же домен или настроен CORS)
    credentials: 'include',
    headers,
  });

  if (res.status === 401) {
    return null;
  }

  if (!res.ok) {
    throw new Error('Network response was not ok');
  }
  const data = (await res.json()) as User;
  return data;
};

export const useCurrentUser = (initialData: User | null) => {
  return useQuery({
    queryKey: ['user', 'currentUser'],
    queryFn: () => fetchCurrentUser(),
    initialData,
  });
};
