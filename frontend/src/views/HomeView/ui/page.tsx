import { fetchPinsPaginated } from '@/entities/pin';
import { Pin } from '@/entities/pin/model/types';
import { ApiPaginatedResponse } from '@/shared/types/response';
import { InfinitePinFeed } from '@/widgets/pin-feed/ui/PinFeed';
import type { Metadata } from 'next';

// ── Метаданные ─────────────────────────────────────────────────

export const metadata: Metadata = {
  title: 'UDPinterest — главная',
  description: 'Лента пинов, вдохновение и идеи',
};

// ── Страница ───────────────────────────────────────────────────

export const HomeView = async () => {
  let initialPages: ApiPaginatedResponse<Pin>[] | undefined;
  try {
    initialPages = [await fetchPinsPaginated()];
  } catch {
    initialPages = [];
  }

  return (
    <main className="min-h-screen pb-10">
      <div className="container mx-auto px-4">
        <InfinitePinFeed initialPages={initialPages} />
      </div>
    </main>
  );
};
