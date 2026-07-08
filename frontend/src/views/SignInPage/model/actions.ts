'use server';

import { z } from 'zod';
import { cookies } from 'next/headers';

// 1. Описываем схему валидации
const loginSchema = z.object({
  email: z.string().email('Введите корректный email'),
  password: z.string().min(5, 'Пароль должен содержать минимум 5 символов'),
});

// Тип для состояния формы
export type AuthState = {
  success?: boolean;
  message?: string;
  errors?: {
    email?: string[];
    password?: string[];
  };
} | null;

// Обрати внимание: теперь первым аргументом идет предыдущее состояние (prevState)
export async function login(prevState: AuthState, formData: FormData): Promise<AuthState> {
  // 2. Преобразуем FormData в обычный объект
  const rawData = Object.fromEntries(formData.entries());

  // 3. Валидируем данные через Zod
  const validatedFields = loginSchema.safeParse(rawData);

  // Если валидация не прошла, возвращаем ошибки конкретных полей
  if (!validatedFields.success) {
    return {
      errors: validatedFields.error.flatten().fieldErrors,
      message: 'Ошибка заполнения формы',
    };
  }

  const { email, password } = validatedFields.data;

  try {
    // 4. Данные валидны, отправляем запрос к Go
    const response = await fetch('http://localhost:8080/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });

    if (!response.ok) {
      return { message: 'Неверный логин или пароль (отказ от Go)' };
    }

    // 5. Устанавливаем куки в Next.js
    const setCookieHeader = response.headers.getSetCookie();
    const cookieStore = await cookies();

    for (const cookieStr of setCookieHeader) {
      const [cookieRaw] = cookieStr.split(';');
      const [name, value] = cookieRaw.split('=');

      cookieStore.set({
        name: name.trim(),
        value: value.trim(),
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        path: '/',
        maxAge: 60 * 60 * 24,
      });
    }

    return { success: true };
  } catch (error) {
    return { message: 'Ошибка соединения с сервером' };
  }
}
