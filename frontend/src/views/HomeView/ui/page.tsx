import { fetchPins } from '@/entities/pin';
import type { Pin } from '@/entities/pin/model/types';
import { PinFeed } from '@/widgets/pin-feed/ui/PinFeed';
import type { Metadata } from 'next';

// ── Метаданные ─────────────────────────────────────────────────

export const metadata: Metadata = {
  title: 'UDPinterest — главная',
  description: 'Лента пинов, вдохновение и идеи',
};

// ── Страница ───────────────────────────────────────────────────

export const HomeView = async () => {
  let initialPins: Pin[] | undefined;
  try {
    initialPins = await fetchPins();
  } catch {
    initialPins = [];
  }

  return (
    <main className="min-h-screen pb-10">
      <div className="container mx-auto px-4">
        <PinFeed initialPins={initialPins} />
      </div>
    </main>
  );
};
