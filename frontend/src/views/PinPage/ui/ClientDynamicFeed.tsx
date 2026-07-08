'use client';

import dynamic from 'next/dynamic';

export const DynamicPinFeed = dynamic(
  () => import('@/widgets/pin-feed/ui/PinFeed').then((mod) => mod.PinFeed),
  {
    ssr: false, // Отрезает компонент от рендеринга на сервере
    loading: () => <div className="animate-pulse">Загрузка сетки фото...</div>, // Что показать вместо него
  },
);
