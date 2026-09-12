package id

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"sync/atomic"
	"time"
)

type idCounter uint32

func (c *idCounter) Increase() uint32 {
	for {
		cur := atomic.LoadUint32((*uint32)(c))
		next := (cur + 1) % 1000
		if atomic.CompareAndSwapUint32((*uint32)(c), cur, next) {
			return cur
		}
	}
}

var orderIdIndex idCounter

// GenerateOrderIdWithRandom 生成订单号，前缀 + 时间戳 + 随机数
func GenerateOrderIdWithRandom(prefix string, tm *time.Time) string {
	// 前缀 + 时间戳（14位） + 随机数（4位）
	// UT3：内联 time 取址，删对根模块 trans 的依赖（该函数本体只是 return &a，
	// 却迫使消费方双 replace bald-utils 根模块 + id 子模块）。

	if tm == nil {
		now := time.Now()
		tm = &now
	}

	timestamp := tm.Format("20060102150405")

	randNum := rand.IntN(10000) // 生成0-9999之间的随机数

	return fmt.Sprintf("%s%s%04d", prefix, timestamp, randNum)
}

// GenerateOrderIdWithIncreaseIndex 生成订单号：前缀 + 时间（14位） + 自增索引（3位补零）。
// UT7 修复：索引 %03d 补零——原 %d 不定长使「20 位订单号」承诺失真；索引 mod 1000
// 环绕，同秒第 1001 单与完全同构的前段 ID 碰撞，使用方须保证单实例同秒 <1000 单
// （多实例须按 worker 分片计数，当前计数器为进程级）。
func GenerateOrderIdWithIncreaseIndex(prefix string, tm *time.Time) string {
	if tm == nil {
		now := time.Now()
		tm = &now
	}

	timestamp := tm.Format("20060102150405")

	index := orderIdIndex.Increase()

	return fmt.Sprintf("%s%s%03d", prefix, timestamp, index)
}

// GenerateOrderIdWithTenantId 带商户ID的订单ID生成器
func GenerateOrderIdWithTenantId(tenantID string) string {
	// 时间戳（14位） + 商户ID（固定 5 位） + 随机数（4位）
	// UT10 修复：按 rune 截断——原 [:5] 按字节截断，中文/emoji 商户名被拦腰
	// 切断产生非法 UTF-8 嵌入订单号。

	// 时间戳部分（精确到毫秒）
	now := time.Now()
	timestamp := now.Format("20060102150405")

	// 商户ID部分（截取或补零到5位）
	tenantPart := tenantID
	if runes := []rune(tenantPart); len(runes) > 5 {
		tenantPart = string(runes[:5])
	} else {
		tenantPart = fmt.Sprintf("%-5s", tenantPart)
		tenantPart = strings.ReplaceAll(tenantPart, " ", "0")
	}

	// 随机数部分（4位）
	n := rand.Int32N(10000)
	randomPart := fmt.Sprintf("%04d", n)

	return timestamp + tenantPart + randomPart
}

func GenerateOrderIdWithPrefixSonyflake(prefix string) string {
	id, _ := NewSonyflakeID()
	return fmt.Sprintf("%s%d", prefix, id)
}

func GenerateOrderIdWithPrefixSnowflake(workerId int64, prefix string) string {
	id, _ := NewSnowflakeID(workerId)
	return fmt.Sprintf("%s%d", prefix, id)
}
