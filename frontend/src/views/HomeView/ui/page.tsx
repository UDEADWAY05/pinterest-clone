import { fetchPins } from '@/entities/pin';
import { PinFeed } from '@/widgets/pin-feed/ui/PinFeed';

// Это компонент-сборщик. Здесь мы будем собирать Хедер, Ленту и Модалки.
export const HomeView = async () => {
  const initialPins = await fetchPins();

  return (
    <main className="min-h-screen pb-10">
      {/* Здесь позже появится <Header /> */}

      {/* Оборачиваем ленту в контейнер */}
      <div className="container mx-auto px-4">
        <PinFeed initialPins={initialPins} />
      </div>
    </main>
  );
};
