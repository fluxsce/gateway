package memory

import (
	"context"
	"fmt"
	"time"
)

// =============================================================================
// 哈希操作的实现
// =============================================================================

// HSet 设置哈希字段
func (m *MemoryCache) HSet(ctx context.Context, key, field, value string) error {
	fullKey := m.buildKey(key)
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	item, exists := m.items[fullKey]
	if !exists || m.isExpired(item) {
		// 创建新的哈希
		hash := make(hashValue)
		hash[field] = value
		
		now := time.Now().UnixNano()
		item = &cacheItem{
			value:       hash,
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
		return nil
	}
	
	// 更新现有哈希
	if hash, ok := item.value.(hashValue); ok {
		hash[field] = value
		item.accessTime = time.Now().UnixNano()
		item.accessCount++
		
		if item.lruNode != nil {
			m.lruList.moveToHead(item.lruNode)
		}
		return nil
	}
	
	return fmt.Errorf("value is not a hash")
}

// HGet 获取哈希字段值
func (m *MemoryCache) HGet(ctx context.Context, key, field string) (string, error) {
	fullKey := m.buildKey(key)
	
	m.mu.RLock()
	item, exists := m.items[fullKey]
	m.mu.RUnlock()
	
	if !exists || m.isExpired(item) {
		return "", nil
	}
	
	if hash, ok := item.value.(hashValue); ok {
		return hash[field], nil
	}
	
	return "", fmt.Errorf("value is not a hash")
}

// HGetAll 获取哈希的所有字段和值
func (m *MemoryCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	fullKey := m.buildKey(key)
	
	m.mu.RLock()
	item, exists := m.items[fullKey]
	m.mu.RUnlock()
	
	if !exists || m.isExpired(item) {
		return make(map[string]string), nil
	}
	
	if hash, ok := item.value.(hashValue); ok {
		result := make(map[string]string)
		for k, v := range hash {
			result[k] = v
		}
		return result, nil
	}
	
	return nil, fmt.Errorf("value is not a hash")
}

// HDel 删除哈希字段
func (m *MemoryCache) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	fullKey := m.buildKey(key)
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	item, exists := m.items[fullKey]
	if !exists || m.isExpired(item) {
		return 0, nil
	}
	
	if hash, ok := item.value.(hashValue); ok {
		deleted := int64(0)
		for _, field := range fields {
			if _, exists := hash[field]; exists {
				delete(hash, field)
				deleted++
			}
		}
		
		item.accessTime = time.Now().UnixNano()
		item.accessCount++
		
		if item.lruNode != nil {
			m.lruList.moveToHead(item.lruNode)
		}
		
		// 如果哈希为空，删除键
		if len(hash) == 0 {
			delete(m.items, fullKey)
			if item.lruNode != nil {
				m.lruList.removeNode(item.lruNode)
			}
		}
		
		return deleted, nil
	}
	
	return 0, fmt.Errorf("value is not a hash")
}

// =============================================================================
// 列表操作的实现
// =============================================================================

// LPush 向列表左侧推入元素
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

// RPush 向列表右侧推入元素
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

// LPop 从列表左侧弹出元素
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

// RPop 从列表右侧弹出元素
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

// LLen 获取列表长度
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

// =============================================================================
// 集合操作的实现
// =============================================================================

// SAdd 向集合添加元素
func (m *MemoryCache) SAdd(ctx context.Context, key string, members ...string) (int64, error) {
	fullKey := m.buildKey(key)
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	item, exists := m.items[fullKey]
	if !exists || m.isExpired(item) {
		// 创建新集合
		set := make(setValue)
		added := int64(0)
		for _, member := range members {
			set[member] = struct{}{}
			added++
		}
		
		now := time.Now().UnixNano()
		item = &cacheItem{
			value:       set,
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
		return added, nil
	}
	
	if set, ok := item.value.(setValue); ok {
		added := int64(0)
		for _, member := range members {
			if _, exists := set[member]; !exists {
				set[member] = struct{}{}
				added++
			}
		}
		
		item.accessTime = time.Now().UnixNano()
		item.accessCount++
		
		if item.lruNode != nil {
			m.lruList.moveToHead(item.lruNode)
		}
		
		return added, nil
	}
	
	return 0, fmt.Errorf("value is not a set")
}

// SRem 从集合移除元素
func (m *MemoryCache) SRem(ctx context.Context, key string, members ...string) (int64, error) {
	fullKey := m.buildKey(key)
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	item, exists := m.items[fullKey]
	if !exists || m.isExpired(item) {
		return 0, nil
	}
	
	if set, ok := item.value.(setValue); ok {
		removed := int64(0)
		for _, member := range members {
			if _, exists := set[member]; exists {
				delete(set, member)
				removed++
			}
		}
		
		item.accessTime = time.Now().UnixNano()
		item.accessCount++
		
		if item.lruNode != nil {
			m.lruList.moveToHead(item.lruNode)
		}
		
		// 如果集合为空，删除键
		if len(set) == 0 {
			delete(m.items, fullKey)
			if item.lruNode != nil {
				m.lruList.removeNode(item.lruNode)
			}
		}
		
		return removed, nil
	}
	
	return 0, fmt.Errorf("value is not a set")
}

// SMembers 获取集合所有成员
func (m *MemoryCache) SMembers(ctx context.Context, key string) ([]string, error) {
	fullKey := m.buildKey(key)
	
	m.mu.RLock()
	item, exists := m.items[fullKey]
	m.mu.RUnlock()
	
	if !exists || m.isExpired(item) {
		return []string{}, nil
	}
	
	if set, ok := item.value.(setValue); ok {
		members := make([]string, 0, len(set))
		for member := range set {
			members = append(members, member)
		}
		return members, nil
	}
	
	return nil, fmt.Errorf("value is not a set")
}

// SIsMember 检查元素是否在集合中
func (m *MemoryCache) SIsMember(ctx context.Context, key string, member string) (bool, error) {
	fullKey := m.buildKey(key)
	
	m.mu.RLock()
	item, exists := m.items[fullKey]
	m.mu.RUnlock()
	
	if !exists || m.isExpired(item) {
		return false, nil
	}
	
	if set, ok := item.value.(setValue); ok {
		_, exists := set[member]
		return exists, nil
	}
	
	return false, fmt.Errorf("value is not a set")
}

// =============================================================================
// 有序集合操作的实现
// =============================================================================

// ZAdd 向有序集合添加元素
func (m *MemoryCache) ZAdd(ctx context.Context, key string, score float64, member string) (int64, error) {
	fullKey := m.buildKey(key)
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	item, exists := m.items[fullKey]
	if !exists || m.isExpired(item) {
		// 创建新有序集合
		zset := make(zsetValue)
		zset[member] = score
		
		now := time.Now().UnixNano()
		item = &cacheItem{
			value:       zset,
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
		return 1, nil
	}
	
	if zset, ok := item.value.(zsetValue); ok {
		added := int64(0)
		if _, exists := zset[member]; !exists {
			added = 1
		}
		zset[member] = score
		
		item.accessTime = time.Now().UnixNano()
		item.accessCount++
		
		if item.lruNode != nil {
			m.lruList.moveToHead(item.lruNode)
		}
		
		return added, nil
	}
	
	return 0, fmt.Errorf("value is not a sorted set")
}

// ZRem 从有序集合移除元素
func (m *MemoryCache) ZRem(ctx context.Context, key string, members ...string) (int64, error) {
	fullKey := m.buildKey(key)
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	item, exists := m.items[fullKey]
	if !exists || m.isExpired(item) {
		return 0, nil
	}
	
	if zset, ok := item.value.(zsetValue); ok {
		removed := int64(0)
		for _, member := range members {
			if _, exists := zset[member]; exists {
				delete(zset, member)
				removed++
			}
		}
		
		item.accessTime = time.Now().UnixNano()
		item.accessCount++
		
		if item.lruNode != nil {
			m.lruList.moveToHead(item.lruNode)
		}
		
		// 如果有序集合为空，删除键
		if len(zset) == 0 {
			delete(m.items, fullKey)
			if item.lruNode != nil {
				m.lruList.removeNode(item.lruNode)
			}
		}
		
		return removed, nil
	}
	
	return 0, fmt.Errorf("value is not a sorted set")
}

// ZScore 获取有序集合成员的分数
func (m *MemoryCache) ZScore(ctx context.Context, key string, member string) (float64, error) {
	fullKey := m.buildKey(key)
	
	m.mu.RLock()
	item, exists := m.items[fullKey]
	m.mu.RUnlock()
	
	if !exists || m.isExpired(item) {
		return 0, nil // 键不存在时返回0，不返回错误
	}
	
	if zset, ok := item.value.(zsetValue); ok {
		if score, exists := zset[member]; exists {
			return score, nil
		}
		return 0, nil // 成员不存在时返回0，不返回错误
	}
	
	return 0, fmt.Errorf("value is not a sorted set")
}

// ZRange 获取有序集合指定范围的成员
func (m *MemoryCache) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	fullKey := m.buildKey(key)
	
	m.mu.RLock()
	item, exists := m.items[fullKey]
	m.mu.RUnlock()
	
	if !exists || m.isExpired(item) {
		return []string{}, nil
	}
	
	if zset, ok := item.value.(zsetValue); ok {
		// 将有序集合转换为排序后的切片
		type memberScore struct {
			member string
			score  float64
		}
		
		memberScores := make([]memberScore, 0, len(zset))
		for member, score := range zset {
			memberScores = append(memberScores, memberScore{member: member, score: score})
		}
		
		// 按分数排序
		for i := 0; i < len(memberScores)-1; i++ {
			for j := i + 1; j < len(memberScores); j++ {
				if memberScores[i].score > memberScores[j].score {
					memberScores[i], memberScores[j] = memberScores[j], memberScores[i]
				}
			}
		}
		
		// 处理范围
		length := int64(len(memberScores))
		if start < 0 {
			start = length + start
		}
		if stop < 0 {
			stop = length + stop
		}
		
		if start < 0 {
			start = 0
		}
		if stop >= length {
			stop = length - 1
		}
		
		if start > stop || start >= length {
			return []string{}, nil
		}
		
		result := make([]string, 0, stop-start+1)
		for i := start; i <= stop; i++ {
			result = append(result, memberScores[i].member)
		}
		
		return result, nil
	}
	
	return nil, fmt.Errorf("value is not a sorted set")
} 