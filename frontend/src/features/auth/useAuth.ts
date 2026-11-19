'use client';

import { useCurrentUser } from '@/entities/user';
import type { User } from '@/entities/user';

/**
 * Хук для проверки аутентификации.
 *
 * @param initialData — опционально, для SSR-гидрации (прокидывается из Server Component)
 */
export function useAuth(initialData?: User | null) {
  const { data: user, isLoading, isError } = useCurrentUser(initialData ?? null);
  return {
    user: user ?? null,
    isAuthenticated: !!user,
    isLoading,
    isError,
  };
}
