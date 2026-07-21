'use client';

import { usePins } from '@/entities/pin';
import { PinCard } from '@/entities/pin';
import { Pin } from '@/entities/pin/model/types';
import Link from 'next/link';

interface PinFeedProps {
  initialPins?: Pin[];
}

export const PinFeed = ({ initialPins }: PinFeedProps) => {
  const { data: pins, isError, isLoading } = usePins(initialPins);

  // Жёсткая загрузка — когда нет ни initialData, ни кеша (без SSR)
  if (isLoading) {
    const getHeightFromIndex = (index: number) => {
      return ((index * 357) % 200) + 150;
    };
    return (
      <div className="columns-2 md:columns-3 lg:columns-4 xl:columns-5 gap-4 max-w-400 mx-auto p-4">
        {Array.from({ length: 10 }).map((_, i) => (
          <div
            key={i}
            className="break-inside-avoid mb-4 rounded-2xl bg-gray-200 animate-pulse"
            style={{ height: getHeightFromIndex(i) }}
          />
        ))}
      </div>
    );
  }

  if (isError && (!pins || pins.length === 0)) {
    return <div className="p-10 text-center text-red-500">Ошибка загрузки данных :(</div>;
  }

  if (!pins || pins.length === 0) {
    return <div className="p-10 text-center">Лента пуста</div>;
  }

  return (
    <div className="columns-2 md:columns-3 lg:columns-4 xl:columns-5 gap-4 max-w-400 mx-auto p-4">
      {pins.map((pin) => (
        <div key={pin.id} className="break-inside-avoid">
          <Link href={`/pin/${pin.id.toString()}`}>
            <PinCard
              height={pin.height}
              dominantColor={pin.dominant_color}
              width={pin.width}
              title={pin.title}
              imageUrl={pin.image_url}
            />
          </Link>
        </div>
      ))}
    </div>
  );
};
