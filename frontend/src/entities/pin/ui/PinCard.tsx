import Image from 'next/image';

interface PinCardProps {
  title: string;
  imageUrl: string;
  width: number;
  height: number;
  dominantColor: string;
  lazy?: boolean;
}

export const PinCard = ({ title, imageUrl, width, height, dominantColor, lazy }: PinCardProps) => {
  // Состояние: загрузилась ли сама картинка в браузере?

  return (
    <div
      className={`relative group rounded-2xl overflow-hidden cursor-pointer mb-4 `}
      // Красим фон в цвет, который прислал бэкенд
      style={{ backgroundColor: dominantColor }}
    >
      <Image
        src={imageUrl}
        alt={title}
        width={width || 300}
        height={height || 200}
        loading={lazy ? 'lazy' : 'eager'}
        // Изначально картинка прозрачная (opacity-0). Как только isImageLoaded станет true, она плавно проявится (opacity-100)
        className={`w-full h-auto object-cover transition-opacity duration-500 ease-in-out opacity-100'`}
        sizes="(max-width: 768px) 100%, (max-width: 1200px) 50vw, 33vw"
      />

      {/* Текст при наведении */}
      <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-end p-4">
        <p className="text-white font-medium truncate">{title}</p>
      </div>
    </div>
  );
};
