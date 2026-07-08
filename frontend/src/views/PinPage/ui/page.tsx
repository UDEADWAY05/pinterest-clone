import { fetchPin } from '@/entities/pin/api/api';
import Image from 'next/image';
import { DynamicPinFeed } from './ClientDynamicFeed';

export async function PinPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const pin = await fetchPin(id);

  if (!pin) return null;
  return (
    <div className="flex flex-col  p-5 justify-center items-center w-full h-fit">
      <div className="w-full border max-w-7xl card rounded-xl items-start border-full overflow-hidden inline-flex flex-row">
        <Image
          src={pin.image_url}
          alt={pin.title}
          width={pin.width}
          height={pin.height}
          className="object-cover w-100 h-100"
        />
        <div className="flex flex-col items-start max-w-100 w-full h-full justify-start p-5">
          <h3 className="text-3xl">{pin.title}</h3>
          <p>{pin.description}</p>
        </div>
      </div>
      <div className="flex flex-wrap">
        <DynamicPinFeed />
      </div>
    </div>
  );
}
