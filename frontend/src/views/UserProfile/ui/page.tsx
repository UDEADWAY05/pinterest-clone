import { fetchUserBoards } from '@/entities/boards';
import { fetchUserById } from '@/entities/user/api/api';
import Image from 'next/image';
import { notFound } from 'next/navigation';
import { Trash2, Pencil } from 'lucide-react';
import './style.css';

type Props = {
  params: Promise<{ id: string }>;
};

export const UserProfilePage = async ({ params }: Props) => {
  const { id } = await params;
  const numId = Number(id);
  if (!id || Number.isNaN(numId)) return notFound();
  const user = await fetchUserById(numId);
  const boards = await fetchUserBoards(numId);
  console.log(boards);

  return (
    <div className="max-w-400 m-auto px-5">
      <div className="p-4 flex">
        <div className="rounded-full w-fit overflow-hidden">
          {user && <Image width={128} height={128} alt="user avatar" src={user.avatar_url ?? ''} />}
        </div>
        <div className="ml-4">
          <h1 className="text-2xl font-semibold">{user?.display_name}</h1>
          <p>
            {user?.followers_count} followers | {user?.following_count} following
          </p>
          <p>{user?.pins_count} pins</p>
          {user?.bio && <p>{user.bio}</p>}
        </div>
      </div>
      {boards && (
        <div className="flex flex-wrap gap-4">
          {boards.map((board) => (
            <div
              className="flex h-48 overflow-hidden rounded-md w-fit group relative duration-300 image-shadow-wrapper"
              key={board.id}
            >
              <div className="absolute flex z-10 bottom-0 left-1 opacity-0 gap-2 rounded-lg px-1 py-1 group-hover:opacity-100 translate-y-2 ease-in-out group-hover:translate-y-0 w-fit h-fit  duration-300">
                <p>{board.title}</p>
              </div>
              <div className="absolute flex top-1 right-1 opacity-0 bg-white/70 gap-2 backdrop-blur-[5px] rounded-lg px-1 py-1 group-hover:opacity-100 -translate-y-2 ease-in-out group-hover:translate-y-0 w-fit h-fit  duration-300">
                <button className="text-zinc-700 hover:bg-gray-400/50 duration-300 px-1 rounded-md">
                  <Pencil className="w-5 h-5" strokeWidth={2} />
                </button>
                <button className="text-zinc-700 hover:bg-gray-400/50 duration-300 p-1 rounded-md">
                  <Trash2 className="w-5 h-5" strokeWidth={2} />
                </button>
              </div>
              <Image
                width={192}
                height={192}
                alt="board cover"
                className="object-cover"
                src={board.recent_pins?.[0]?.image_url ?? ''}
              />
              <div className="flex flex-col overflow-hidden">
                <Image
                  width={128}
                  height={96}
                  objectFit="cover"
                  alt="board cover"
                  className="object-cover h-24 w-24"
                  src={board.recent_pins?.[1]?.image_url ?? ''}
                />
                <Image
                  width={128}
                  height={96}
                  alt="board cover"
                  className="object-cover h-24 w-24"
                  src={board.recent_pins?.[2]?.image_url ?? ''}
                />
              </div>
            </div>
          ))}
        </div>
      )}
      {user?.is_current_user && <p>Ваши пины</p>}
    </div>
  );
};
