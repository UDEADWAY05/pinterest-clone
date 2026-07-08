'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { useState } from 'react';

export function QueryProvider({ children }: { children: React.ReactNode }) {
  // Создаем QueryClient внутри useState, чтобы он не пересоздавался 
  // при каждом ререндере компонента в React
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 60 * 1000, // Данные свежие 1 минуту
            refetchOnWindowFocus: false, // Не делать запрос, если юзер переключил вкладку
          },
        },
      })
  );

  return (
    <QueryClientProvider client={queryClient}>
      {children}
      {/* Девтулзы будут видны только в режиме разработки (иконка цветка в углу) */}
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
}