import { ApiResponse } from '@/shared/types/response';
import { BoardWithRecentPins } from './types';
import { cookies } from 'next/headers';

const BACKEND_URL = process.env.NEXT_PUBLIC_BACKEND_URL ?? 'http://localhost:8080';

const baseUrl = typeof window === 'undefined' ? BACKEND_URL : '';

const REVALIDATE_SECONDS = 60;

export const fetchUserBoards = async (id: number): Promise<BoardWithRecentPins[] | null> => {
  try {
    const cookieStore = await cookies();
    const sessionId = cookieStore.get('session')?.value;

    const headers: Record<string, string> = {};

    if (sessionId) {
      headers['cookie'] = `session=${sessionId}`;
    }
    const res = await fetch(`${baseUrl}/api/users/${id}/boards`, {
      next: { revalidate: REVALIDATE_SECONDS },
      headers,
    });
    if (res.status === 404) return null;
    if (!res.ok) throw new Error(`Failed to fetch boards`);
    return await res.json();
  } catch (error) {
    console.error(error);
    return null;
  }
};
