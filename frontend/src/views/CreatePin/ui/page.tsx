'use client';
import { useState, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { presignUpload, uploadFile, getImageDimensions, useCreatePin } from '../model/api';
import Image from 'next/image';
import { useForm } from 'react-hook-form';
import z from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';

const MAX_FILE_SIZE = 5 * 1024 * 1024; // 5MB
const ACCEPTED_IMAGE_TYPES = ['image/jpeg', 'image/jpg', 'image/png', 'image/webp'];

const schema = z.object({
  file: z
    .custom<File>((val) => val instanceof File, 'Загрузите файл')
    .refine((file) => file?.size <= MAX_FILE_SIZE, 'Максимальный размер — 5MB.')
    .refine(
      (file) => ACCEPTED_IMAGE_TYPES.includes(file?.type),
      'Поддерживаются только .jpg, .jpeg, .png и .webp.',
    ),
  title: z.string().min(1, 'Название обязательно'),
  description: z.string().optional(),
  tags: z.string().optional(),
});

type CreatePinFormData = z.infer<typeof schema>;

export const CreatePin = () => {
  const router = useRouter();
  const fileRef = useRef<HTMLInputElement>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const {
    handleSubmit: onSubmit,
    register,
    setValue,
    formState: { errors },
  } = useForm<CreatePinFormData>({
    defaultValues: {
      title: '',
      description: '',
      tags: '',
    },
    resolver: zodResolver(schema),
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0];
    if (!f) return;
    setValue('file', f);
    setPreview(URL.createObjectURL(f));
  };

  const { mutateAsync: createPin } = useCreatePin();

  const handleSubmit = async (data: CreatePinFormData) => {
    if (!data.file || !data.title.trim()) return;
    setIsSubmitting(true);
    setError(null);

    try {
      // 1. Получаем presigned URL
      const presign = await presignUpload({
        file_name: data.file.name,
        content_type: data.file.type,
        file_size: data.file.size,
      });

      // 2. Загружаем файл в S3
      await uploadFile(presign.upload_url, data.file);

      // 3. Читаем размеры изображения
      const { width, height } = await getImageDimensions(data.file);

      // 4. Создаём пин
      const pin = await createPin({
        title: data.title.trim(),
        description: data.description?.trim() || '',
        image_url: presign.public_url,
        width,
        height,
        dominant_color: '#e5e7eb', // fallback — бэкенд может пересчитать
        tags: data.tags
          ?.split(',')
          .map((t) => t.trim())
          .filter(Boolean),
      });

      router.push(`/pin/${pin.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка создания пина');
    } finally {
      setIsSubmitting(false);
    }
  };

  // const canSubmit = file && title.trim() && !isSubmitting;

  return (
    <main className="min-h-screen flex items-start justify-center p-6 pt-12">
      <form onSubmit={onSubmit(handleSubmit)} className="w-full max-w-2xl space-y-6">
        <h1 className="text-3xl font-bold">Создать пин</h1>

        {/* Зона загрузки изображения */}
        <div
          onClick={() => fileRef.current?.click()}
          className={`relative aspect-square w-full max-w-sm mx-auto rounded-2xl border-2 border-dashed cursor-pointer
            flex items-center justify-center overflow-hidden bg-gray-100
            ${preview ? 'border-transparent' : 'border-gray-300 hover:border-gray-400'}`}
        >
          {preview ? (
            <Image src={preview} alt="preview" fill className="object-cover" />
          ) : (
            <span className="text-gray-400 text-lg">Нажмите, чтобы выбрать изображение</span>
          )}
          <input
            type="file"
            accept="image/jpeg,image/png,image/webp"
            ref={fileRef}
            onChange={handleFileSelect}
          />
        </div>
        {errors.file && <span className="text-red-500 text-sm">{errors.file.message}</span>}

        {/* Поля формы */}
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Название</label>
            <Input placeholder="Добавьте название..." {...register('title')} />
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Описание</label>
            <textarea
              {...register('description')}
              placeholder="Добавьте описание..."
              rows={3}
              className="flex w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm
                shadow-sm placeholder:text-muted-foreground focus-visible:outline-none
                focus-visible:ring-1 focus-visible:ring-ring resize-none"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Теги (через запятую)</label>
            <Input {...register('tags')} placeholder="например: природа, горы, путешествия" />
          </div>
        </div>

        {error && <p className="text-sm text-red-500 font-medium">{error}</p>}

        <Button type="submit" className="w-full" size="lg">
          {isSubmitting ? 'Загрузка...' : 'Опубликовать'}
        </Button>
      </form>
    </main>
  );
};
