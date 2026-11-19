'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import { useInfinitePins, PinCard } from '@/entities/pin';
import { PaginatedPins, Pin } from '@/entities/pin/model/types';
import Link from 'next/link';

interface InfinitePinFeedProps {
  initialPages?: PaginatedPins[];
}

const getHeightFromIndex = (index: number) => ((index * 357) % 200) + 150;

function useColumnCount() {
  const [columnCount, setColumnCount] = useState(5);

  useEffect(() => {
    const updateColumns = () => {
      const width = window.innerWidth;
      if (width < 768) setColumnCount(2);
      else if (width < 1024) setColumnCount(3);
      else if (width < 1280) setColumnCount(4);
      else setColumnCount(5);
    };

    updateColumns();
    window.addEventListener('resize', updateColumns);
    return () => window.removeEventListener('resize', updateColumns);
  }, []);

  return columnCount;
}

const PinSkeletonGrid = ({ count }: { count: number }) => (
  <div className="columns-2 md:columns-3 lg:columns-4 xl:columns-5 gap-4 max-w-400 mx-auto p-4">
    {Array.from({ length: count }).map((_, i) => (
      <div
        key={i}
        className="break-inside-avoid mb-4 rounded-2xl bg-gray-200 animate-pulse"
        style={{ height: getHeightFromIndex(i) }}
      />
    ))}
  </div>
);

export const InfinitePinFeed = ({ initialPages }: InfinitePinFeedProps) => {
  const { data, isError, isLoading, isFetchingNextPage, hasNextPage, fetchNextPage } =
    useInfinitePins(initialPages);

  const columnCount = useColumnCount();

  const sentinelRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel || !hasNextPage || isFetchingNextPage) return;

    const observer = new IntersectionObserver((entries) => {
      if (entries[0].isIntersecting) {
        fetchNextPage();
      }
    });
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  const pins = useMemo(() => {
    return data?.pages.flatMap((page) => page.data) ?? [];
  }, [data]);

  // 2. Стабильное распределение пинов по фиксированным колонкам
  const columns = useMemo(() => {
    const cols: Pin[][] = Array.from({ length: columnCount }, () => []);
    pins.forEach((pin, index) => {
      cols[index % columnCount].push(pin);
    });
    return cols;
  }, [pins, columnCount]);

  if (isLoading && pins.length === 0) {
    return <PinSkeletonGrid count={6} />;
  }

  if (isError && pins.length === 0) {
    return <div className="p-10 text-center text-red-500">Ошибка загрузки данных :(</div>;
  }

  if (pins.length === 0) {
    return <div className="p-10 text-center">Лента пуста</div>;
  }

  return (
    <div className="max-w-400 mx-auto p-4">
      <div className="flex gap-4 items-start">
        {columns.map((colPins, colIdx) => (
          <div key={colIdx} className="flex flex-col gap-4 flex-1 min-w-0">
            {colPins.map((pin) => (
              <div key={pin.id}>
                <Link href={`/pin/${pin.id.toString()}`}>
                  <PinCard
                    lazy
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
        ))}
      </div>
      <div ref={sentinelRef} className="h-1" aria-hidden />
      {isFetchingNextPage && <PinSkeletonGrid count={6} />}
    </div>
  );
};
