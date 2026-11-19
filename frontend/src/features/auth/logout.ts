'use server';

import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';

const BACKEND_URL = process.env.NEXT_PUBLIC_BACKEND_URL ?? 'http://localhost:8080';

export async function logout() {
  const cookieStore = await cookies();
  const sessionCookie = cookieStore.get('session');

  if (sessionCookie?.value) {
    // Инвалидируем сессию на бэкенде
    await fetch(`${BACKEND_URL}/api/auth/logout`, {
      method: 'POST',
      headers: { cookie: `session=${sessionCookie.value}` },
    });
  }

  // Удаляем куку session в Next.js
  cookieStore.delete('session');

  redirect('/signin');
}
