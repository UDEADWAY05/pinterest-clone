import { useInfiniteQuery, useQuery } from '@tanstack/react-query';
import { Pin } from '../model/types';
import { ApiPaginatedResponse } from '@/shared/types/response';

const BACKEND_URL = process.env.NEXT_PUBLIC_BACKEND_URL ?? 'http://localhost:8080';

const baseUrl = typeof window === 'undefined' ? BACKEND_URL : '';

const REVALIDATE_SECONDS = 60;
const STALE_TIME = 5 * 60 * 1000;
const PAGE_SIZE = 10;

export const fetchPinsPaginated = async (
  cursor?: string,
  limit: number = PAGE_SIZE,
): Promise<ApiPaginatedResponse<Pin>> => {
  const params = new URLSearchParams({ limit: String(limit) });
  if (cursor) params.set('cursor', cursor);
  const res = await fetch(`${baseUrl}/api/pins?${params.toString()}`, {
    next: { revalidate: REVALIDATE_SECONDS },
  });
  if (!res.ok) throw new Error('Failed to fetch pins');
  return await res.json();
};

export const fetchPin = async (id: string): Promise<Pin | null> => {
  const res = await fetch(`${baseUrl}/api/pins/${id}`, {
    next: { revalidate: REVALIDATE_SECONDS },
  });
  if (res.status === 404) return null;
  if (!res.ok) throw new Error(`Failed to fetch pin ${id}`);
  return await res.json();
};

export const useInfinitePins = (initialPages?: ApiPaginatedResponse<Pin>[]) => {
  return useInfiniteQuery({
    queryKey: ['pins', 'feed'],
    queryFn: ({ pageParam }) => fetchPinsPaginated(pageParam),
    initialPageParam: undefined as string | undefined,
    initialData: initialPages
      ? { pages: initialPages, pageParams: initialPages.map(() => undefined as string | undefined) }
      : undefined,
    getNextPageParam: (lastPage) => lastPage.next_cursor ?? undefined,
    staleTime: STALE_TIME,
  });
};

export const usePin = (id: string, initialData?: Pin) => {
  return useQuery({
    queryKey: ['pin', id],
    queryFn: () => fetchPin(id),
    initialData,
    staleTime: STALE_TIME,
  });
};
