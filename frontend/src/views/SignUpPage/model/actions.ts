'use server';

import { z } from 'zod';
import { cookies } from 'next/headers';

const signUpSchema = z.object({
  email: z.string().trim().email('Введите корректный email'),
  password: z.string().min(5, 'Пароль должен содержать минимум 5 символов'),
  display_name: z.string().min(2, 'Имя должно содержать минимум 2 символа'),
  avatar_url: z.string().optional(),
});

export type SignUpState = {
  success?: boolean;
  message?: string;
  errors?: {
    email?: string[];
    password?: string[];
    display_name?: string[];
    avatar_url?: string[];
  };
} | null;

export async function signUp(prevState: SignUpState, formData: FormData): Promise<SignUpState> {
  const rawData = Object.fromEntries(formData.entries());

  const validatedFields = signUpSchema.safeParse(rawData);

  if (!validatedFields.success) {
    return {
      errors: validatedFields.error.flatten().fieldErrors,
      message: 'Ошибка заполнения формы',
    };
  }

  const { email, password, display_name, avatar_url } = validatedFields.data;

  try {
    const response = await fetch('http://localhost:8080/api/auth/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email,
        password,
        display_name,
        avatar_url: avatar_url || '',
      }),
    });

    if (!response.ok) {
      const body = await response.json().catch(() => null);
      const message = body?.error?.message || 'Ошибка регистрации';
      return { message };
    }

    // Пробрасываем Set-Cookie от Go в Next.js
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
        maxAge: 60 * 60 * 24 * 7,
      });
    }

    return { success: true };
  } catch {
    return { message: 'Ошибка соединения с сервером' };
  }
}
