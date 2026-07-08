'use client';

import { usePins } from '@/entities/pin';
import { PinCard } from '@/entities/pin';
import { Pin } from '@/entities/pin/model/types';
import Link from 'next/link';

interface PinFeedProps {
  initialPins?: Pin[];
}

export const PinFeed = ({ initialPins }: PinFeedProps) => {
  // Передаем initialPins в наш хук.
  // Теперь isLoading ВСЕГДА будет false при первой загрузке!
  const { data: pins, isError } = usePins(initialPins);

  if (isError) {
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
