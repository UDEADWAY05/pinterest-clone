import { ApiResponse } from '@/shared/types/response';
import { CreatePinRequest, PresignRequest, PresignResult } from './types';
import { Pin } from '@/entities/pin/model/types';
import { useMutation, useQueryClient } from '@tanstack/react-query';

/** Получить presigned URL для загрузки изображения. */

export async function presignUpload(req: PresignRequest): Promise<PresignResult> {
  const res = await fetch('/api/uploads/presign', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(req),
  });
  if (!res.ok) throw new Error('Failed to get upload URL');
  const body: ApiResponse<PresignResult> = await res.json();
  return body.data;
}

/** Загрузить файл в S3 по presigned URL. */
export async function uploadFile(uploadUrl: string, file: File): Promise<void> {
  const res = await fetch(uploadUrl, {
    method: 'PUT',
    body: file,
    headers: { 'Content-Type': file.type },
  });
  if (!res.ok) throw new Error('Failed to upload file');
}

/** Прочитать размеры изображения из File. */
export function getImageDimensions(file: File): Promise<{ width: number; height: number }> {
  return new Promise((resolve, reject) => {
    const url = URL.createObjectURL(file);
    const img = new Image();
    img.onload = () => {
      URL.revokeObjectURL(url);
      resolve({ width: img.naturalWidth, height: img.naturalHeight });
    };
    img.onerror = () => {
      URL.revokeObjectURL(url);
      reject(new Error('Failed to load image'));
    };
    img.src = url;
  });
}

export async function createPinRequest(req: CreatePinRequest): Promise<Pin> {
  const res = await fetch('/api/pins', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const errBody = await res.json().catch(() => null);
    throw new Error(errBody?.error?.message ?? 'Failed to create pin');
  }
  const body: ApiResponse<Pin> = await res.json();
  return body.data;
}

export const useCreatePin = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (req: CreatePinRequest) => createPinRequest(req),
    onSuccess: () => {
      // Автоматически инвалидируем ленту пинов после создания
      queryClient.invalidateQueries({ queryKey: ['pins', 'feed'] });
    },
  });
};
