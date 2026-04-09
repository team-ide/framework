package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"time"
)

func NewLocker(service IService) *Locker {
	res := new(Locker)
	res.service = service
	return res
}

type Locker struct {
	service IService
}

func (this_ *Locker) Lock(ctx context.Context, lockKey string, timeout time.Duration) (string, error) {
	client, err := this_.service.GetClient()
	if err != nil {
		return "", err
	}
	token, _, err := acquireLock(ctx, client, lockKey, timeout)
	return token, err
}
func (this_ *Locker) Unlock(ctx context.Context, lockKey string, token string) error {
	client, err := this_.service.GetClient()
	if err != nil {
		return err
	}
	_, err = releaseLock(ctx, client, lockKey, token)
	return err
}

// 尝试获取锁
func acquireLock(ctx context.Context, client redis.Cmdable, lockKey string, ttl time.Duration) (string, bool, error) {
	// 生成一个唯一的 token，用于标识当前客户端
	token := uuid.New().String()

	// 使用 SetNX 命令，它对应 Redis 的 SET key value NX EX seconds
	// 这保证了 "检查是否不存在" 和 "设置过期时间" 是一个原子操作
	success, err := client.SetNX(ctx, lockKey, token, ttl).Result()
	if err != nil {
		return "", false, err
	}
	return token, success, nil
}

// 定义解锁的 Lua 脚本
var releaseLockScript = redis.NewScript(`
    if redis.call("get", KEYS[1]) == ARGV[1] then
        return redis.call("del", KEYS[1])
    else
        return 0
    end
`)

// 释放锁
func releaseLock(ctx context.Context, client redis.Cmdable, lockKey string, token string) (bool, error) {
	// 通过 EvalSha 或 Run 方法执行 Lua 脚本
	// Keys: [lockKey], Args: [token]
	res, err := releaseLockScript.Run(ctx, client, []string{lockKey}, token).Result()
	if err != nil {
		return false, err
	}
	// 脚本返回 1 表示删除成功，0 表示值不匹配或键不存在
	return res == int64(1), nil
}
