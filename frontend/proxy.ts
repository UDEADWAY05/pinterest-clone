import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function proxy(request: NextRequest) {
  const sessionToken = request.cookies.get('session')?.value;
  const { pathname } = request.nextUrl;

  // Если куки нет, редиректим на /login
  if (!sessionToken) {
    const loginUrl = new URL('/signin', request.url);
    loginUrl.searchParams.set('from', pathname);
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
}

// Задаем matcher ТОЛЬКО для защищенных роутов
export const config = {
  matcher: [
    '/create/:path*',
    // '/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
  extends: '/api/*',
};
