### UDPinterest

#### Описание
Это клон популярного сервиса под названием Pinterest

#### Установка и запуск
1. Клонируйте репозиторий: `git clone https://github.com/yourusername/udpinterest.git`
2. Перейдите в директорию проекта: `cd udpinterest`
3. Перейдите в директорию `backend`: `cd backend`
4. go mod tidy
5. Запустите docker compose: `docker-compose up -d`
6. Запустите сервер: `go run main.go`
7. Перейдите в директорию `frontend`: `cd frontend`
8. Обновите .env файл
9. Запустите фронтенд: `npm install`
10. Запустите фронтенд: `npm start`

#### Технологии
##### Backend
- Go
- Redis
- PostgreSQL

##### Frontend
- React
- Next.js
- Tailwind CSS
- Tanstack Query
- Shadcn UI
