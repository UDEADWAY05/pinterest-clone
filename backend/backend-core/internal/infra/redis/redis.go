package redis

import (
	"context"
	"fmt"
	"time"

	"pinterest-core/internal/config"

	goredis "github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *goredis.Client
}

func New(ctx context.Context, cfg *config.Config) (*Redis, error) {
	// 1. Настройка параметров подключения (Connection Pool)
	opts := &goredis.Options{
		Addr:         cfg.RedisAddr,
		PoolSize:     100,             // Максимальное количество соединений
		MinIdleConns: 10,              // Минимальное количество "спящих" соединений
		DialTimeout:  5 * time.Second, // Таймаут на установку соединения
		ReadTimeout:  3 * time.Second, // Таймаут на чтение
		WriteTimeout: 3 * time.Second, // Таймаут на запись
	}

	// 2. Инициализация клиента
	client := goredis.NewClient(opts)

	// 3. Проверка доступности (Fail Fast)
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close() // Обязательно закрываем клиент, если пинг не прошел
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Redis{Client: client}, nil
}

func (r *Redis) Close(context.Context) error {
	if r == nil || r.Client == nil {
		return nil
	}

	// Закрываем все соединения в пуле
	return r.Client.Close()
}
