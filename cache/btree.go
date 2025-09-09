package cache

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type PersistMode bool

// BTree 表示B+树KV存储系统
type BTree struct {
	filePath          string
	file              *os.File
	walFile           *os.File // WAL文件
	mu                sync.RWMutex
	root              *node
	order             int
	pageSize          int
	persistMode       PersistMode   // 持久化模式: realtime(实时)或timed(定时)
	persistInterval   time.Duration // 定时持久化间隔
	dirty             bool          // 数据是否被修改，用于增量模式
	maxKeySize        int           // key最大长度
	updatedKeys       []string      // 定时模式下更新的键列表
	walRecordCount    int           // WAL记录计数器
	walMergeThreshold int           // WAL合并阈值
}

// node 表示B+树节点
type node struct {
	isLeaf   bool
	keys     []string
	values   [][]byte
	children []*node
	expires  []int64 // 过期时间戳，0表示永不过期
	next     *node   // 叶子节点的下一个节点
}

// KVEntry 表示键值对条目
type KVEntry struct {
	Key    string
	Value  []byte
	Expire int64
}

const (
	pageSize               = 4096
	headerSize             = 128
	maxKeySize             = 200   // key最大长度200字节
	Realtime   PersistMode = false // 实时持久化模式
	Timed      PersistMode = true  // 定时持久化模式
)

// NewBTree 创建新的B+树实例默认实时持久化模式
func NewBTree(filePath string) (*BTree, error) {
	return NewBTreeWithConfig(filePath, Realtime, 0)
}

// NewBTreeWithConfig 创建新的B+树实例（支持配置持久化模式）
func NewBTreeWithConfig(filePath string, persistMode PersistMode, persistInterval time.Duration) (*BTree, error) {
	return NewBTreeWithAdvancedConfig(filePath, persistMode, persistInterval, 1000)
}

// NewBTreeWithAdvancedConfig 创建新的B+树实例（支持高级配置）
func NewBTreeWithAdvancedConfig(filePath string, persistMode PersistMode, persistInterval time.Duration, walMergeThreshold int) (*BTree, error) {
	bt := &BTree{
		filePath:          filePath,
		pageSize:          pageSize,
		order:             100, // 根据页面大小自动计算
		persistMode:       persistMode,
		persistInterval:   persistInterval,
		maxKeySize:        maxKeySize,
		dirty:             false,
		updatedKeys:       make([]string, 0),
		walRecordCount:    0,
		walMergeThreshold: walMergeThreshold, // 用户可配置的WAL合并阈值
	}

	err := bt.init()
	if err != nil {
		return nil, err
	}

	return bt, nil
}

// init 初始化B+树
func (bt *BTree) init() error {
	var err error

	// 打开或创建文件
	bt.file, err = os.OpenFile(bt.filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	// 打开或创建WAL文件
	walPath := bt.filePath + ".wal"
	bt.walFile, err = os.OpenFile(walPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	// 读取文件头信息
	info, err := bt.file.Stat()
	if err != nil {
		return err
	}

	if info.Size() == 0 {
		// 新文件，写入文件头
		err = bt.writeHeader()
		if err != nil {
			return err
		}
		bt.root = bt.newLeafNode()
	} else {
		// 从文件加载数据
		err = bt.loadFromFile()
		if err != nil {
			return err
		}
	}

	// 启动定时持久化
	go bt.startAutoPersist()

	return nil
}

// copyActiveKeysToTempFile 从原数据文件复制活跃键值对到临时文件
func (bt *BTree) copyActiveKeysToTempFile(tempFile *os.File, activeKeys map[string]bool) error {
	// 重置数据文件指针到文件头之后
	if _, err := bt.file.Seek(int64(headerSize), 0); err != nil {
		return err
	}

	// 遍历数据文件中的所有键值对
	for {
		header := make([]byte, 16)
		_, err := bt.file.Read(header)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		keyLen := int(binary.BigEndian.Uint32(header[0:4]))
		valLen := int(binary.BigEndian.Uint32(header[4:8]))

		// 读取键数据
		keyBytes := make([]byte, keyLen)
		if _, err := bt.file.Read(keyBytes); err != nil {
			return err
		}
		key := string(keyBytes)

		// 读取值数据
		value := make([]byte, valLen)
		if _, err := bt.file.Read(value); err != nil {
			return err
		}

		// 检查该键是否活跃（未被删除且是最新版本）
		if _, exists := activeKeys[key]; exists {
			// 如果WAL中有这个键的记录，跳过原数据文件中的旧版本
			continue
		}

		// 写入键值对到临时文件（只有未被WAL覆盖的键才需要复制）
		if _, err := tempFile.Write(header); err != nil {
			return err
		}
		if _, err := tempFile.Write(keyBytes); err != nil {
			return err
		}
		if _, err := tempFile.Write(value); err != nil {
			return err
		}
	}

	return nil
}

// writeHeader 写入文件头
func (bt *BTree) writeHeader() error {
	header := make([]byte, headerSize)
	binary.BigEndian.PutUint32(header[0:4], uint32(bt.order))
	binary.BigEndian.PutUint32(header[4:8], uint32(bt.pageSize))
	_, err := bt.file.WriteAt(header, 0)
	return err
}

// writeWALRecord 写入WAL记录
func (bt *BTree) writeWALRecord(entry KVEntry, operation byte) error {
	// 序列化WAL记录: 操作类型(1字节) + 键长度(4字节) + 值长度(4字节) + 过期时间(8字节) + 键数据 + 值数据
	keyBytes := []byte(entry.Key)
	keyLen := len(keyBytes)
	valLen := len(entry.Value)

	record := make([]byte, 1+4+4+8+keyLen+valLen)
	record[0] = operation
	binary.BigEndian.PutUint32(record[1:5], uint32(keyLen))
	binary.BigEndian.PutUint32(record[5:9], uint32(valLen))
	binary.BigEndian.PutUint64(record[9:17], uint64(entry.Expire))

	copy(record[17:17+keyLen], keyBytes)
	copy(record[17+keyLen:17+keyLen+valLen], entry.Value)

	_, err := bt.walFile.Write(record)
	return err
}

// applyWALToDataFile 应用WAL到数据文件
func (bt *BTree) applyWALToDataFile() error {
	// 重置WAL文件指针到开头
	if _, err := bt.walFile.Seek(0, 0); err != nil {
		return err
	}

	// 创建临时文件用于日志结构合并
	tempFilePath := bt.filePath + ".tmp"
	tempFile, err := os.OpenFile(tempFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer tempFile.Close()
	defer os.Remove(tempFilePath)

	// 写入文件头到临时文件
	header := make([]byte, headerSize)
	binary.BigEndian.PutUint32(header[0:4], uint32(bt.order))
	binary.BigEndian.PutUint32(header[4:8], uint32(bt.pageSize))
	if _, err = tempFile.Write(header); err != nil {
		return err
	}

	// 读取数据文件中的所有有效键值对（跳过已删除的键）
	activeKeys := make(map[string]bool)
	var tempOffset int64 = headerSize

	// 读取并处理所有WAL记录
	for {
		header := make([]byte, 17)
		_, err = bt.walFile.Read(header)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		operation := header[0]
		keyLen := int(binary.BigEndian.Uint32(header[1:5]))
		valLen := int(binary.BigEndian.Uint32(header[5:9]))
		expire := int64(binary.BigEndian.Uint64(header[9:17]))

		// 读取键数据
		keyBytes := make([]byte, keyLen)
		if _, err = bt.walFile.Read(keyBytes); err != nil {
			return err
		}
		key := string(keyBytes)

		// 读取值数据
		value := make([]byte, valLen)
		if _, err = bt.walFile.Read(value); err != nil {
			return err
		}

		// 根据操作类型更新活跃键集合
		switch operation {
		case 'P': // Put操作
			activeKeys[key] = true

			// 写入键值对到临时文件
			kvHeader := make([]byte, 16)
			binary.BigEndian.PutUint32(kvHeader[0:4], uint32(keyLen))
			binary.BigEndian.PutUint32(kvHeader[4:8], uint32(valLen))
			binary.BigEndian.PutUint64(kvHeader[8:16], uint64(expire))

			// 写入头部
			if _, err = tempFile.Write(kvHeader); err != nil {
				return err
			}
			tempOffset += 16

			// 写入键
			if _, err = tempFile.Write(keyBytes); err != nil {
				return err
			}
			tempOffset += int64(keyLen)

			// 写入值
			if _, err = tempFile.Write(value); err != nil {
				return err
			}
			tempOffset += int64(valLen)
		case 'D': // Delete操作
			activeKeys[key] = false
		}
	}

	// 从原数据文件中复制未被删除的键值对
	if err = bt.copyActiveKeysToTempFile(tempFile, activeKeys); err != nil {
		return err
	}

	// 用临时文件替换原数据文件
	if err = tempFile.Sync(); err != nil {
		return err
	}
	if err = bt.file.Close(); err != nil {
		return err
	}
	if err = os.Rename(tempFilePath, bt.filePath); err != nil {
		return err
	}

	// 重新打开数据文件
	bt.file, err = os.OpenFile(bt.filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	// 清空WAL文件
	if err := bt.walFile.Truncate(0); err != nil {
		return err
	}
	if _, err := bt.walFile.Seek(0, 0); err != nil {
		return err
	}

	return nil
}

// newLeafNode 创建新的叶子节点
func (bt *BTree) newLeafNode() *node {
	return &node{
		isLeaf:  true,
		keys:    make([]string, 0),
		values:  make([][]byte, 0),
		expires: make([]int64, 0),
	}
}

// startAutoPersist 启动自动持久化定时器
func (bt *BTree) startAutoPersist() {
	// 只有定时模式才启动定时器
	if bt.persistMode != Timed {
		return
	}

	ticker := time.NewTicker(bt.persistInterval)
	defer ticker.Stop()

	for range ticker.C {
		bt.mu.Lock()
		if bt.dirty {
			bt.persistToFile()
			bt.dirty = false
		}
		bt.mu.Unlock()
	}
}

// Put 插入键值对
func (bt *BTree) Put(key string, value []byte, ttl time.Duration) error {
	// 验证key长度
	if len(key) > bt.maxKeySize {
		return fmt.Errorf("key长度超过限制(%d字节)", bt.maxKeySize)
	}

	bt.mu.Lock()
	defer bt.mu.Unlock()

	var expire int64
	if ttl > 0 {
		expire = time.Now().Add(ttl).Unix()
	}

	// 根据持久化模式处理记录
	if bt.persistMode == Realtime {
		// 实时模式：写入WAL记录
		walEntry := KVEntry{Key: key, Value: value, Expire: expire}
		if err := bt.writeWALRecord(walEntry, 'P'); err != nil {
			return err
		}
		bt.walRecordCount++
	} else {
		// 定时模式：记录更新的键
		bt.updatedKeys = append(bt.updatedKeys, key)
	}

	err := bt.insert(key, value, expire)
	if err != nil {
		return err
	}

	// 标记数据已修改
	bt.dirty = true

	// 实时模式：当WAL记录达到阈值时才持久化
	if bt.persistMode == Realtime && bt.walRecordCount >= bt.walMergeThreshold {
		if err := bt.persistToFile(); err != nil {
			return err
		}
		bt.dirty = false
		bt.walRecordCount = 0
	}

	return nil
}

// Get 获取键值对
func (bt *BTree) Get(key string) ([]byte, error) {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	return bt.search(key)
}

// Delete 删除键值对
func (bt *BTree) Delete(key string) error {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	// 根据持久化模式处理记录
	if bt.persistMode == Realtime {
		// 实时模式：写入WAL记录
		walEntry := KVEntry{Key: key, Value: nil, Expire: 0}
		if err := bt.writeWALRecord(walEntry, 'D'); err != nil {
			return err
		}
		bt.walRecordCount++
	} else {
		// 定时模式：记录删除的键
		bt.updatedKeys = append(bt.updatedKeys, key)
	}

	err := bt.delete(key)
	if err != nil {
		return err
	}

	// 标记数据已修改
	bt.dirty = true

	// 实时模式：当WAL记录达到阈值时才持久化
	if bt.persistMode == Realtime && bt.walRecordCount >= bt.walMergeThreshold {
		if err := bt.persistToFile(); err != nil {
			return err
		}
		bt.dirty = false
		bt.walRecordCount = 0
	}

	return nil
}

// BulkPut 批量插入键值对（增量模式）
func (bt *BTree) BulkPut(entries []KVEntry) error {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	// 验证所有key长度
	for _, entry := range entries {
		if len(entry.Key) > bt.maxKeySize {
			return fmt.Errorf("key '%s' 长度超过限制(%d字节)", entry.Key, bt.maxKeySize)
		}
		if bt.persistMode == Realtime {
			// 实时模式：写入WAL记录
			if err := bt.writeWALRecord(entry, 'P'); err != nil {
				return err
			}
			bt.walRecordCount++
		} else {
			// 定时模式：记录更新的键
			bt.updatedKeys = append(bt.updatedKeys, entry.Key)
		}

		if err := bt.insert(entry.Key, entry.Value, entry.Expire); err != nil {
			return err
		}
	}
	bt.dirty = true

	// 实时模式：当WAL记录达到阈值时才持久化
	if bt.persistMode == Realtime && bt.walRecordCount >= bt.walMergeThreshold {
		if err := bt.persistToFile(); err != nil {
			return err
		}
		bt.dirty = false
		bt.walRecordCount = 0
	}

	return nil
}

// BulkGet 批量查询键值对
func (bt *BTree) BulkGet(keys []string) ([]KVEntry, error) {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	results := make([]KVEntry, 0, len(keys))
	for _, key := range keys {
		value, err := bt.search(key)
		if err == nil {
			results = append(results, KVEntry{Key: key, Value: value})
		}
	}

	return results, nil
}

// CleanExpired 清理过期键值对
func (bt *BTree) CleanExpired() int {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	return bt.cleanExpired()
}

// Count 统计键值对数量
func (bt *BTree) Count() int {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	return bt.count()
}

// Flush 清空数据库
func (bt *BTree) Flush() error {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	bt.root = bt.newLeafNode()

	// 清空数据文件
	if err := bt.file.Truncate(headerSize); err != nil {
		return err
	}

	// 清空WAL文件
	if bt.walFile != nil {
		if err := bt.walFile.Truncate(0); err != nil {
			return err
		}
		if _, err := bt.walFile.Seek(0, 0); err != nil {
			return err
		}
	}

	return nil
}

// List 列出所有键值对
func (bt *BTree) List() ([]KVEntry, error) {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	if bt.root == nil {
		return []KVEntry{}, nil
	}

	entries := make([]KVEntry, 0)
	current := bt.root

	// 找到第一个叶子节点
	for !current.isLeaf {
		current = current.children[0]
	}

	// 遍历所有叶子节点
	for current != nil {
		for i := 0; i < len(current.keys); i++ {
			// 跳过过期键
			if current.expires[i] > 0 && time.Now().Unix() > current.expires[i] {
				continue
			}
			entries = append(entries, KVEntry{
				Key:    current.keys[i],
				Value:  current.values[i],
				Expire: current.expires[i],
			})
		}
		current = current.next
	}

	return entries, nil
}

// Keys 获取所有键
func (bt *BTree) Keys() ([]string, error) {
	entries, err := bt.List()
	if err != nil {
		return nil, err
	}

	keys := make([]string, len(entries))
	for i, entry := range entries {
		keys[i] = entry.Key
	}

	return keys, nil
}

// Values 获取所有值
func (bt *BTree) Values() ([][]byte, error) {
	entries, err := bt.List()
	if err != nil {
		return nil, err
	}

	values := make([][]byte, len(entries))
	for i, entry := range entries {
		values[i] = entry.Value
	}

	return values, nil
}

// Close 关闭B+树
func (bt *BTree) Close() error {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	// 最后持久化（无论WAL记录数量多少，关闭时都要持久化）
	if bt.dirty {
		if err := bt.persistToFile(); err != nil {
			return err
		}
	}

	// 关闭WAL文件
	if bt.walFile != nil {
		if err := bt.walFile.Close(); err != nil {
			return err
		}
	}

	return bt.file.Close()
}

// insert 插入键值对到B+树
func (bt *BTree) insert(key string, value []byte, expire int64) error {
	// 实现B+树插入逻辑
	if bt.root == nil {
		bt.root = bt.newLeafNode()
	}

	// 简单的叶子节点插入（简化实现）
	if bt.root.isLeaf {
		return bt.insertIntoLeaf(bt.root, key, value, expire)
	}

	// 非叶子节点需要递归插入
	return bt.insertIntoNode(bt.root, key, value, expire)
}

// insertIntoLeaf 插入到叶子节点
func (bt *BTree) insertIntoLeaf(leaf *node, key string, value []byte, expire int64) error {
	// 查找插入位置
	pos := 0
	for pos < len(leaf.keys) && leaf.keys[pos] < key {
		pos++
	}

	// 如果键已存在，更新值
	if pos < len(leaf.keys) && leaf.keys[pos] == key {
		leaf.values[pos] = value
		leaf.expires[pos] = expire
		return nil
	}

	// 插入新键值对
	leaf.keys = append(leaf.keys[:pos], append([]string{key}, leaf.keys[pos:]...)...)
	leaf.values = append(leaf.values[:pos], append([][]byte{value}, leaf.values[pos:]...)...)
	leaf.expires = append(leaf.expires[:pos], append([]int64{expire}, leaf.expires[pos:]...)...)

	// 检查是否需要分裂
	if len(leaf.keys) > bt.order {
		return bt.splitLeaf(leaf)
	}

	return nil
}

// search 搜索键值对
func (bt *BTree) search(key string) ([]byte, error) {
	if bt.root == nil {
		return nil, fmt.Errorf("key not found")
	}

	current := bt.root
	for !current.isLeaf {
		// 在非叶子节点中查找正确的分支
		pos := 0
		for pos < len(current.keys) && current.keys[pos] <= key {
			pos++
		}
		// 直接使用pos作为索引，因为children比keys多一个
		current = current.children[pos]
	}

	// 在叶子节点中查找
	for i, k := range current.keys {
		if k == key {
			// 检查是否过期
			if current.expires[i] > 0 && time.Now().Unix() > current.expires[i] {
				return nil, fmt.Errorf("key expired")
			}
			return current.values[i], nil
		}
	}

	return nil, fmt.Errorf("key not found")
}

// delete 删除键值对
func (bt *BTree) delete(key string) error {
	if bt.root == nil {
		return fmt.Errorf("key not found")
	}

	// 简化实现：在叶子节点中删除
	if bt.root.isLeaf {
		return bt.deleteFromLeaf(bt.root, key)
	}

	// 非叶子节点需要递归删除
	return bt.deleteFromNode(bt.root, key)
}

// deleteFromLeaf 从叶子节点删除
func (bt *BTree) deleteFromLeaf(leaf *node, key string) error {
	for i, k := range leaf.keys {
		if k == key {
			// 删除键值对
			leaf.keys = append(leaf.keys[:i], leaf.keys[i+1:]...)
			leaf.values = append(leaf.values[:i], leaf.values[i+1:]...)
			leaf.expires = append(leaf.expires[:i], leaf.expires[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("key not found")
}

// cleanExpired 清理过期键值对
func (bt *BTree) cleanExpired() int {
	if bt.root == nil {
		return 0
	}

	count := 0
	current := bt.root

	// 找到第一个叶子节点
	for !current.isLeaf {
		current = current.children[0]
	}

	// 遍历所有叶子节点
	for current != nil {
		for i := 0; i < len(current.keys); i++ {
			if current.expires[i] > 0 && time.Now().Unix() > current.expires[i] {
				// 删除过期键值对
				current.keys = append(current.keys[:i], current.keys[i+1:]...)
				current.values = append(current.values[:i], current.values[i+1:]...)
				current.expires = append(current.expires[:i], current.expires[i+1:]...)
				count++
				i-- // 调整索引
			}
		}
		current = current.next
	}

	return count
}

// count 统计键值对数量
func (bt *BTree) count() int {
	if bt.root == nil {
		return 0
	}

	count := 0
	current := bt.root

	// 找到第一个叶子节点
	for !current.isLeaf {
		current = current.children[0]
	}

	// 遍历所有叶子节点计数
	for current != nil {
		count += len(current.keys)
		current = current.next
	}

	return count
}

// persistToFile 持久化到文件
func (bt *BTree) persistToFile() error {
	if bt.file == nil {
		return fmt.Errorf("文件未打开")
	}

	if bt.persistMode == Realtime {
		// 实时模式：使用WAL文件
		// 检查WAL文件是否有内容
		walInfo, err := bt.walFile.Stat()
		if err != nil {
			return err
		}

		if walInfo.Size() == 0 {
			// WAL为空，无需持久化
			return nil
		}

		// 应用WAL记录到数据文件
		if err := bt.applyWALToDataFile(); err != nil {
			return err
		}

		// 清空WAL文件
		if err := bt.walFile.Truncate(0); err != nil {
			return err
		}
		if _, err := bt.walFile.Seek(0, 0); err != nil {
			return err
		}

		// 重置WAL记录计数器
		bt.walRecordCount = 0
	} else {
		// 定时模式：使用updatedKeys进行增量更新
		if len(bt.updatedKeys) == 0 {
			// 没有更新的极，无需持久化
			return nil
		}

		// 创建临时文件进行增量更新
		if err := bt.applyUpdatedKeysToDataFile(); err != nil {
			return err
		}

		// 清空更新的键列表
		bt.updatedKeys = make([]string, 0)
	}

	return bt.file.Sync()
}

// applyUpdatedKeysToDataFile 应用更新的键到数据文件
func (bt *BTree) applyUpdatedKeysToDataFile() error {
	// 创建临时文件用于增量更新
	tempFilePath := bt.filePath + ".tmp"
	tempFile, err := os.OpenFile(tempFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer tempFile.Close()
	defer os.Remove(tempFilePath)

	// 写入文件头到临时文件
	header := make([]byte, headerSize)
	binary.BigEndian.PutUint32(header[0:4], uint32(bt.order))
	binary.BigEndian.PutUint32(header[4:8], uint32(bt.pageSize))
	if _, err = tempFile.Write(header); err != nil {
		return err
	}

	// 从原数据文件中复制未被更新的键值对
	if err = bt.copyUnupdatedKeysToTempFile(tempFile); err != nil {
		return err
	}

	// 写入更新的键值对到临时文件
	if err = bt.writeUpdatedKeysToTempFile(tempFile); err != nil {
		return err
	}

	// 用临时文件替换原数据文件
	if err = tempFile.Sync(); err != nil {
		return err
	}
	if err = bt.file.Close(); err != nil {
		return err
	}

	if err = os.Rename(tempFilePath, bt.filePath); err != nil {
		return err
	}

	// 重新打开数据文件
	bt.file, err = os.OpenFile(bt.filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	return nil
}

// copyUnupdatedKeysToTempFile 复制未被更新的键值对到临时文件
func (bt *BTree) copyUnupdatedKeysToTempFile(tempFile *os.File) error {
	if _, err := bt.file.Seek(int64(headerSize), 0); err != nil {
		return err
	}

	updatedKeysSet := make(map[string]bool)
	for _, key := range bt.updatedKeys {
		updatedKeysSet[key] = true
	}

	// 读取原数据文件中的所有键值对
	for {
		// 读取键值对头
		header := make([]byte, 16)
		if _, err := bt.file.Read(header); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		keyLen := int(binary.BigEndian.Uint32(header[0:4]))
		valLen := int(binary.BigEndian.Uint32(header[4:8]))
		_ = int64(binary.BigEndian.Uint64(header[8:16])) // expire字段，这里不需要使用

		// 读取键数据
		keyBytes := make([]byte, keyLen)
		if _, err := bt.file.Read(keyBytes); err != nil {
			return err
		}
		key := string(keyBytes)

		// 读取值数据
		value := make([]byte, valLen)
		if _, err := bt.file.Read(value); err != nil {
			return err
		}

		// 如果这个键没有被更新，则复制到临时文件
		if !updatedKeysSet[key] {
			// 写入键值对头
			if _, err := tempFile.Write(header); err != nil {
				return err
			}
			// 写入键数据
			if _, err := tempFile.Write(keyBytes); err != nil {
				return err
			}
			// 写入值数据
			if _, err := tempFile.Write(value); err != nil {
				return err
			}
		}
	}

	return nil
}

// writeUpdatedKeysToTempFile 写入更新的键值对到临时文件
func (bt *BTree) writeUpdatedKeysToTempFile(tempFile *os.File) error {
	updatedKeysSet := make(map[string]bool)
	for _, key := range bt.updatedKeys {
		updatedKeysSet[key] = true
	}

	// 遍历B+树中的所有键值对，只写入被更新的键
	if bt.root == nil {
		return nil
	}

	current := bt.root
	// 找到第一个叶子节点
	for !current.isLeaf {
		current = current.children[0]
	}

	// 遍历所有叶子节点
	for current != nil {
		for i, key := range current.keys {
			if updatedKeysSet[key] {
				// 写入键值对头
				header := make([]byte, 16)
				binary.BigEndian.PutUint32(header[0:4], uint32(len(key)))
				binary.BigEndian.PutUint32(header[4:8], uint32(len(current.values[i])))
				binary.BigEndian.PutUint64(header[8:16], uint64(current.expires[i]))
				if _, err := tempFile.Write(header); err != nil {
					return err
				}

				// 写入键数据
				if _, err := tempFile.Write([]byte(key)); err != nil {
					return err
				}
				// 写入值数据
				if _, err := tempFile.Write(current.values[i]); err != nil {
					return err
				}
			}
		}
		current = current.next
	}

	return nil
}

// loadFromFile 从文件加载数据
func (bt *BTree) loadFromFile() error {
	if bt.file == nil {
		return fmt.Errorf("文件未打开")
	}

	// 读取文件头
	header := make([]byte, headerSize)
	if _, err := bt.file.ReadAt(header, 0); err != nil {
		return err
	}

	bt.order = int(binary.BigEndian.Uint32(header[0:4]))
	bt.pageSize = int(binary.BigEndian.Uint32(header[4:8]))

	// 重建B+树
	bt.root = bt.newLeafNode()
	offset := headerSize

	for {
		// 读取键值对头
		header := make([]byte, 16)
		if _, err := bt.file.ReadAt(header, int64(offset)); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		offset += 16

		keyLen := int(binary.BigEndian.Uint32(header[0:4]))
		valLen := int(binary.BigEndian.Uint32(header[4:8]))
		expire := int64(binary.BigEndian.Uint64(header[8:16]))

		// 读取键数据
		keyBytes := make([]byte, keyLen)
		if _, err := bt.file.ReadAt(keyBytes, int64(offset)); err != nil {
			return err
		}
		offset += keyLen
		key := string(keyBytes)

		// 读取值数据
		value := make([]byte, valLen)
		if _, err := bt.file.ReadAt(value, int64(offset)); err != nil {
			return err
		}
		offset += valLen

		// 插入到B+树
		if err := bt.insert(key, value, expire); err != nil {
			return err
		}
	}

	// 检查并应用WAL文件
	walInfo, err := bt.walFile.Stat()
	if err != nil {
		return err
	}

	if walInfo.Size() > 0 {
		// 应用WAL记录
		if err := bt.applyWALToDataFile(); err != nil {
			return err
		}
	}

	return nil
}

// splitLeaf 分裂叶子节点
func (bt *BTree) splitLeaf(leaf *node) error {
	// 创建新的叶子节点
	newLeaf := bt.newLeafNode()

	// 计算分裂点（取中间位置）
	mid := len(leaf.keys) / 2

	// 将后半部分数据移动到新节点
	newLeaf.keys = append(newLeaf.keys, leaf.keys[mid:]...)
	newLeaf.values = append(newLeaf.values, leaf.values[mid:]...)
	newLeaf.expires = append(newLeaf.expires, leaf.expires[mid:]...)

	// 更新原节点的数据
	leaf.keys = leaf.keys[:mid]
	leaf.values = leaf.values[:mid]
	leaf.expires = leaf.expires[:mid]

	// 更新叶子节点链表
	newLeaf.next = leaf.next
	leaf.next = newLeaf

	// 如果这是根节点，需要创建新的根节点
	if leaf == bt.root {
		newRoot := &node{
			isLeaf:   false,
			keys:     []string{newLeaf.keys[0]},
			children: []*node{leaf, newLeaf},
		}
		bt.root = newRoot
		return nil
	}

	// 否则需要将新节点插入到父节点中
	// 使用新节点的第一个键作为分隔键
	return bt.insertIntoParent(leaf, newLeaf.keys[0], newLeaf)
}

// insertIntoNode 插入到非叶子节点
func (bt *BTree) insertIntoNode(n *node, key string, value []byte, expire int64) error {
	// 查找合适的子节点
	pos := 0
	for pos < len(n.keys) && n.keys[pos] <= key {
		pos++
	}

	// 递归插入到子节点
	child := n.children[pos]
	if child.isLeaf {
		return bt.insertIntoLeaf(child, key, value, expire)
	}
	return bt.insertIntoNode(child, key, value, expire)
}

// deleteFromNode 从非叶子节点删除
func (bt *BTree) deleteFromNode(n *node, key string) error {
	// 查找合适的子节点
	pos := 0
	for pos < len(n.keys) && n.keys[pos] <= key {
		pos++
	}

	// 递归删除
	child := n.children[pos]
	if child.isLeaf {
		return bt.deleteFromLeaf(child, key)
	}
	return bt.deleteFromNode(child, key)
}

// insertIntoParent 将新节点插入到父节点中
func (bt *BTree) insertIntoParent(oldNode *node, key string, newNode *node) error {
	// 使用BFS查找父节点，避免递归深度问题
	if bt.root == oldNode {
		// 如果没有父节点，说明oldNode是根节点，需要创建新的根节点
		newRoot := &node{
			isLeaf:   false,
			keys:     []string{key},
			children: []*node{oldNode, newNode},
		}
		bt.root = newRoot
		return nil
	}

	// 使用队列进行BFS查找父节点
	queue := []*node{bt.root}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.isLeaf {
			continue
		}

		// 检查当前节点的所有子节点
		for _, child := range current.children {
			if child == oldNode {
				// 找到父节点
				parent := current

				// 查找插入位置
				pos := 0
				for pos < len(parent.keys) && parent.keys[pos] < key {
					pos++
				}

				// 插入新键和子节点
				parent.keys = append(parent.keys[:pos], append([]string{key}, parent.keys[pos:]...)...)
				parent.children = append(parent.children[:pos+1], append([]*node{newNode}, parent.children[pos+1:]...)...)

				// 检查是否需要分裂非叶子节点
				if len(parent.keys) > bt.order {
					return bt.splitNonLeaf(parent)
				}
				return nil
			}

			// 将非叶子节点加入队列继续搜索
			if !child.isLeaf {
				queue = append(queue, child)
			}
		}
	}

	// 如果没有找到父节点，说明oldNode是根节点，需要创建新的根节点
	newRoot := &node{
		isLeaf:   false,
		keys:     []string{key},
		children: []*node{oldNode, newNode},
	}
	bt.root = newRoot
	return nil
}

// splitNonLeaf 分裂非叶子节点
func (bt *BTree) splitNonLeaf(n *node) error {
	// 创建新的非叶子节点
	newNode := &node{
		isLeaf:   false,
		keys:     make([]string, 0),
		children: make([]*node, 0),
	}

	// 计算分裂点
	mid := len(n.keys) / 2
	midKey := n.keys[mid]

	// 将后半部分数据移动到新节点
	newNode.keys = append(newNode.keys, n.keys[mid+1:]...)
	newNode.children = append(newNode.children, n.children[mid+1:]...)

	// 更新原节点的数据
	n.keys = n.keys[:mid]
	n.children = n.children[:mid+1]

	// 如果这是根节点，需要创建新的根节点
	if n == bt.root {
		newRoot := &node{
			isLeaf:   false,
			keys:     []string{midKey},
			children: []*node{n, newNode},
		}
		bt.root = newRoot
		return nil
	}

	// 否则需要将新节点插入到父节点中
	return bt.insertIntoParent(n, midKey, newNode)
}

// SetWALMergeThreshold 设置WAL合并阈值
func (bt *BTree) SetWALMergeThreshold(threshold int) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	bt.walMergeThreshold = threshold
}

// GetWALRecordCount 获取当前WAL记录数量
func (bt *BTree) GetWALRecordCount() int {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	return bt.walRecordCount
}

// ForceWALMerge 强制立即合并WAL文件
func (bt *BTree) ForceWALMerge() error {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if bt.persistMode == Realtime && bt.walRecordCount > 0 {
		return bt.persistToFile()
	}
	return nil
}
