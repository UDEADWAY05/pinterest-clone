import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  /* config options here */
  reactCompiler: true,
  images: {
    dangerouslyAllowLocalIP: true,
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'i.pinimg.com', // Разрешаем скачивать картинки с этого домена
        port: '',
        pathname: '/**', // Разрешаем любые пути после домена
      },
      {
        protocol: 'https',
        hostname: 'avatars.fastly.steamstatic.com', // Разрешаем скачивать картинки с этого домена
        port: '',
        pathname: '/**', // Разрешаем любые пути после домена
      },
      {
        protocol: 'http',
        hostname: 'localhost',
        port: '9000',
        pathname: '/**',
      },
      {
        protocol: 'http',
        hostname: '127.0.0.1',
        port: '9000',
        pathname: '/**',
      },
      // {
      //   protocol: 'http',
      //   hostname: 'localhost', // Разрешаем скачивать картинки с этого домена
      //   port: '8080',
      //   pathname: '/**', // Разрешаем любые пути после домена
      // },
    ],
  },
  async rewrites() {
    return [{ source: '/api/:path*', destination: 'http://localhost:8080/api/:path*' }];
  },
};

export default nextConfig;
