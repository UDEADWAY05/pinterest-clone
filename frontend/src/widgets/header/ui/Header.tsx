import { fetchCurrentUser } from '@/entities/user/api/api';
import { Bell, MessageCircle, Search, User } from 'lucide-react';
import Link from 'next/link';
import { cookies } from 'next/headers';
import Image from 'next/image';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/shared/ui/dropdown-menu';

export const Header = async () => {
  const currentUser = await fetchCurrentUser().catch(() => null);

  return (
    <header className="fixed top-0 left-0 right-0 h-20 z-50 py-4 px-6 flex items-center gap-2">
      <Link href="/" className="h-12 text-[1.2rem] flex items-center justify-center">
        Pinterest
      </Link>

      <Link
        href="/create"
        className="px-4 py-2 border glass-container border-[rgba(255,255,255,0.15)] glass  bg-[rgba(255,255,255,0.05)] hover:bg-gray-100/15 rounded-full font-medium text-base transition-colors"
      >
        Создать
        <div className="glass-container-child"></div>
      </Link>

      {/* 3. Строка поиска (пока только UI) */}
      <div className="flex-1 flex items-center glass glass-container transition-colors rounded-full px-4 py-2 h-12">
        <Search className="w-5 h-5 mr-2" />
        <input
          type="text"
          placeholder="Поиск"
          className="bg-transparent outline-none w-full text-base"
        />
      </div>

      {/* 4. Меню пользователя (Иконки) */}
      <div className="flex items-center gap-1">
        <button className="w-12 h-12 rounded-full hover:bg-gray-100 dark:hover:bg-gray-100/10 flex p-0 items-center justify-center transition-colors text-white">
          <Bell className="w-6 h-6" />
        </button>
        <button className="w-12 h-12 rounded-full hover:bg-gray-100 dark:hover:bg-gray-100/10 flex p-0 items-center justify-center transition-colors text-white">
          <MessageCircle className="w-6 h-6" />
        </button>
        {currentUser ? (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button className="w-12 h-12 rounded-full hover:bg-gray-100 dark:hover:bg-gray-100/10 flex p-0 items-center justify-center transition-colors text-white">
                <Image
                  width={40}
                  height={40}
                  src={currentUser.avatar_url}
                  alt={currentUser.display_name}
                  className="w-8 h-8 rounded-full"
                />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <Link href={`${currentUser.id}`}>
                <DropdownMenuItem>Profile</DropdownMenuItem>
              </Link>
            </DropdownMenuContent>
          </DropdownMenu>
        ) : (
          <Link
            href={'/signin'}
            className="w-12 h-12 rounded-full hover:bg-gray-100 dark:hover:bg-gray-100/10 flex p-0 items-center justify-center transition-colors text-white"
          >
            <User className="w-6 h-6" />
          </Link>
        )}
      </div>
    </header>
  );
};
