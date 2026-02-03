package memory

import (
	"context"
	"fmt"
	"time"
)

// =============================================================================
// 列表操作
// =============================================================================

// LPush 向列表左侧推入元素。
//
// 在列表的左侧（头部）插入一个或多个值。
// 如果列表不存在，会自动创建。
//
// 参数：
//   - ctx: 上下文
//   - key: 列表键名
//   - values: 要插入的值列表
//
// 返回值：
//   - int64: 操作后列表的长度
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	// 单个元素
//	length, err := cache.LPush(ctx, "queue", "task1")
//
//	// 多个元素（从左到右依次插入）
//	length, err := cache.LPush(ctx, "queue", "task1", "task2", "task3")
func (m *MemoryCache) LPush(ctx context.Context, key string, values ...string) (int64, error) {
	fullKey := m.buildKey(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.items[fullKey]
	if !exists || m.isExpired(item) {
		// 创建新列表
		list := make(listValue, 0, len(values))
		// 左侧推入，需要反序添加
		for i := len(values) - 1; i >= 0; i-- {
			list = append(list, values[i])
		}

		now := time.Now().UnixNano()
		item = &cacheItem{
			value:       list,
			expiration:  m.resolveExpiration(m.config.DefaultExpiration),
			accessTime:  now,
			accessCount: 1,
			createTime:  now,
		}

		if m.config.EvictionPolicy == EvictionLRU {
			item.lruNode = &lruNode{key: fullKey}
			m.lruList.addToHead(item.lruNode)
		}

		m.items[fullKey] = item
		return int64(len(list)), nil
	}

	if list, ok := item.value.(listValue); ok {
		// 左侧推入，需要在前面插入
		newList := make(listValue, 0, len(list)+len(values))
		// 先添加新值（反序）
		for i := len(values) - 1; i >= 0; i-- {
			newList = append(newList, values[i])
		}
		// 再添加原有值
		newList = append(newList, list...)

		item.value = newList
		item.accessTime = time.Now().UnixNano()
		item.accessCount++

		if item.lruNode != nil {
			m.lruList.moveToHead(item.lruNode)
		}

		return int64(len(newList)), nil
	}

	return 0, fmt.Errorf("value is not a list")
}

// RPush 向列表右侧推入元素。
//
// 在列表的右侧（尾部）插入一个或多个值。
//
// 参数：
//   - ctx: 上下文
//   - key: 列表键名
//   - values: 要插入的值列表
//
// 返回值：
//   - int64: 操作后列表的长度
//   - error: 操作失败时返回错误
func (m *MemoryCache) RPush(ctx context.Context, key string, values ...string) (int64, error) {
	fullKey := m.buildKey(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.items[fullKey]
	if !exists || m.isExpired(item) {
		// 创建新列表
		list := make(listValue, len(values))
		copy(list, values)

		now := time.Now().UnixNano()
		item = &cacheItem{
			value:       list,
			expiration:  m.resolveExpiration(m.config.DefaultExpiration),
			accessTime:  now,
			accessCount: 1,
			createTime:  now,
		}

		if m.config.EvictionPolicy == EvictionLRU {
			item.lruNode = &lruNode{key: fullKey}
			m.lruList.addToHead(item.lruNode)
		}

		m.items[fullKey] = item
		return int64(len(list)), nil
	}

	if list, ok := item.value.(listValue); ok {
		list = append(list, values...)
		item.value = list
		item.accessTime = time.Now().UnixNano()
		item.accessCount++

		if item.lruNode != nil {
			m.lruList.moveToHead(item.lruNode)
		}

		return int64(len(list)), nil
	}

	return 0, fmt.Errorf("value is not a list")
}

// LPop 从列表左侧弹出元素。
//
// 从列表的左侧（头部）弹出并返回一个元素。
// 如果弹出后列表为空，会自动删除整个键。
//
// 参数：
//   - ctx: 上下文
//   - key: 列表键名
//
// 返回值：
//   - string: 弹出的元素值，列表为空时返回空字符串
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	task, err := cache.LPop(ctx, "queue")
//	if err != nil {
//	    return err
//	}
//	if task == "" {
//	    log.Info("队列为空")
//	}
func (m *MemoryCache) LPop(ctx context.Context, key string) (string, error) {
	fullKey := m.buildKey(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.items[fullKey]
	if !exists || m.isExpired(item) {
		return "", nil
	}

	if list, ok := item.value.(listValue); ok {
		if len(list) == 0 {
			return "", nil
		}

		value := list[0]
		list = list[1:]
		item.value = list
		item.accessTime = time.Now().UnixNano()
		item.accessCount++

		if item.lruNode != nil {
			m.lruList.moveToHead(item.lruNode)
		}

		// 如果列表为空，删除键
		if len(list) == 0 {
			delete(m.items, fullKey)
			if item.lruNode != nil {
				m.lruList.removeNode(item.lruNode)
			}
		}

		return value, nil
	}

	return "", fmt.Errorf("value is not a list")
}

// RPop 从列表右侧弹出元素。
//
// 从列表的右侧（尾部）弹出并返回一个元素。
//
// 参数：
//   - ctx: 上下文
//   - key: 列表键名
//
// 返回值：
//   - string: 弹出的元素值，列表为空时返回空字符串
//   - error: 操作失败时返回错误
func (m *MemoryCache) RPop(ctx context.Context, key string) (string, error) {
	fullKey := m.buildKey(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.items[fullKey]
	if !exists || m.isExpired(item) {
		return "", nil
	}

	if list, ok := item.value.(listValue); ok {
		if len(list) == 0 {
			return "", nil
		}

		lastIndex := len(list) - 1
		value := list[lastIndex]
		list = list[:lastIndex]
		item.value = list
		item.accessTime = time.Now().UnixNano()
		item.accessCount++

		if item.lruNode != nil {
			m.lruList.moveToHead(item.lruNode)
		}

		// 如果列表为空，删除键
		if len(list) == 0 {
			delete(m.items, fullKey)
			if item.lruNode != nil {
				m.lruList.removeNode(item.lruNode)
			}
		}

		return value, nil
	}

	return "", fmt.Errorf("value is not a list")
}

// LLen 获取列表长度。
//
// 返回列表的元素数量。
//
// 参数：
//   - ctx: 上下文
//   - key: 列表键名
//
// 返回值：
//   - int64: 列表长度，列表不存在时返回 0
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	length, err := cache.LLen(ctx, "queue")
//	log.Info("队列长度:", length)
func (m *MemoryCache) LLen(ctx context.Context, key string) (int64, error) {
	fullKey := m.buildKey(key)

	m.mu.RLock()
	item, exists := m.items[fullKey]
	m.mu.RUnlock()

	if !exists || m.isExpired(item) {
		return 0, nil
	}

	if list, ok := item.value.(listValue); ok {
		return int64(len(list)), nil
	}

	return 0, fmt.Errorf("value is not a list")
}
