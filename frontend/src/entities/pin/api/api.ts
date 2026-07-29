import { useQuery } from '@tanstack/react-query';
import { Pin } from '../model/types';
import type { ApiResponse } from '@/shared/types/response';

const BACKEND_URL = process.env.NEXT_PUBLIC_BACKEND_URL ?? 'http://localhost:8080';

const baseUrl = typeof window === 'undefined' ? BACKEND_URL : '';

const REVALIDATE_SECONDS = 60;

export const fetchPins = async (): Promise<Pin[]> => {
  const res = await fetch(`${baseUrl}/api/pins`, {
    next: { revalidate: REVALIDATE_SECONDS },
  });
  if (!res.ok) throw new Error('Failed to fetch pins');
  const body: ApiResponse<Pin[]> = await res.json();
  return body.data;
};

export const fetchPin = async (id: string): Promise<Pin | null> => {
  const res = await fetch(`${baseUrl}/api/pins/${id}`, {
    next: { revalidate: REVALIDATE_SECONDS },
  });
  if (res.status === 404) return null;
  if (!res.ok) throw new Error(`Failed to fetch pin ${id}`);
  const body: ApiResponse<Pin> = await res.json();
  return body.data;
};

const STALE_TIME = 5 * 60 * 1000;

export const usePins = (initialData?: Pin[]) => {
  return useQuery({
    queryKey: ['pins', 'feed'],
    queryFn: fetchPins,
    initialData,
    staleTime: STALE_TIME,
  });
};
