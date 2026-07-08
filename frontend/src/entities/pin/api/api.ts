import { useQuery } from '@tanstack/react-query';
import { Pin } from '../model/types';

const BACKEND_URL = process.env.NEXT_PUBLIC_BACKEND_URL ?? 'http://localhost:8080';
// Эту функцию мы будем использовать и на сервере, и на клиенте
export const fetchPins = async (): Promise<Pin[]> => {
  const res = await fetch(`${typeof window === 'undefined' ? BACKEND_URL : ''}/api/pins`, {
    // cache: "no-cache",
    // next: { revalidate: 15 },
  });
  if (!res.ok) throw new Error('Network response was not ok');
  return res.json();
};

export const fetchPin = async (id: string): Promise<Pin> => {
  const res = await fetch(`${typeof window === 'undefined' ? BACKEND_URL : ''}/api/pins/${id}`, {
    // cache: "no-cache",
    // next: { revalidate: 15 },
  });
  if (!res.ok) throw new Error('Network response was not ok');
  return res.json();
};

// Добавляем параметр initialData
export const usePins = (initialData?: Pin[]) => {
  return useQuery({
    queryKey: ['pins', 'feed'],
    queryFn: fetchPins,
    initialData: initialData,
  });
};
