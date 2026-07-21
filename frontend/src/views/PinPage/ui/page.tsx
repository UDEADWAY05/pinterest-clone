import { fetchPin } from '@/entities/pin/api/api';
import Image from 'next/image';
import { notFound } from 'next/navigation';
import type { Metadata } from 'next';
import { DynamicPinFeed } from './ClientDynamicFeed';

interface Props {
  params: Promise<{ id: string }>;
}
export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { id } = await params;
  const pin = await fetchPin(id);
  if (!pin) return { title: 'Пин не найден' };

  return {
    title: pin.title,
    description: pin.description ?? '',
    openGraph: {
      title: pin.title,
      description: pin.description ?? '',
      images: [pin.image_url],
    },
  };
}

export default async function PinPage({ params }: Props) {
  const { id } = await params;
  const pin = await fetchPin(id);

  if (!pin) notFound();

  return (
    <div className="flex flex-col p-5 justify-center items-center w-full h-fit">
      <div className="w-full border max-w-7xl card rounded-xl items-start border-full overflow-hidden inline-flex flex-row">
        <Image
          src={pin.image_url}
          alt={pin.title}
          width={pin.width}
          height={pin.height}
          priority
          className="object-cover w-100 h-100"
        />
        <div className="flex flex-col items-start max-w-100 w-full h-full justify-start p-5">
          <h1 className="text-3xl">{pin.title}</h1>
          <p>{pin.description}</p>
        </div>
      </div>
      <div className="flex flex-wrap">
        <DynamicPinFeed />
      </div>
    </div>
  );
}
