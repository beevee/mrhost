package redis

import (
	"encoding/json"
	"fmt"

	"mrhost"

	"github.com/go-redis/redis"
)

func getChatMetaKey(chatID int64) string {
	return fmt.Sprintf("chat-metas:%d", chatID)
}

func (r Repository) GetChatMeta(chatID int64) (meta mrhost.ChatMeta, err error) {
	result, err := r.client.Get(getChatMetaKey(chatID)).Result()
	if err == redis.Nil {
		return meta, nil
	}
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(result), &meta)
	return
}

func (r Repository) SetChatMeta(chatID int64, meta mrhost.ChatMeta) error {
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	_, err = r.client.Set(getChatMetaKey(chatID), metaJSON, 0).Result()
	if err != nil {
		return err
	}
	return nil
}
