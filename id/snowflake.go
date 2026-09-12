package id

import (
	"sync"

	"github.com/bwmarrin/snowflake"
)

var snowflakeNodeMap = sync.Map{}

type SnowflakeNode struct {
	workerId int64
	node     *snowflake.Node
	sync.Mutex
}

func NewSnowflakeNode(workerId int64) (*SnowflakeNode, error) {
	node, err := snowflake.NewNode(workerId)
	return &SnowflakeNode{
		workerId: workerId,
		node:     node,
		Mutex:    sync.Mutex{},
	}, err
}

func (sfNode *SnowflakeNode) Generate() int64 {
	sfNode.Lock()
	defer sfNode.Unlock()
	return sfNode.node.Generate().Int64()
}

func (sfNode *SnowflakeNode) GenerateString() string {
	sfNode.Lock()
	defer sfNode.Unlock()
	return sfNode.node.Generate().String()
}

func NewSnowflakeID(workerId int64) (int64, error) {
	// 64 位 ID = 41 位时间戳 + 10 位工作节点 ID + 12 位序列号
	//
	// UT2 修复：LoadOrStore 原子化 check-then-act——并发 miss 时两 goroutine
	// 各建同 workerId 的 Node、后写覆盖先写、败者节点仍被本地引用继续出号
	// （snowflake 去重只在单 Node 内部，双 Node 并发可产出重复 int64）。
	// 败者丢弃自建节点，统一用先入库的节点。
	if find, ok := snowflakeNodeMap.Load(workerId); ok {
		return find.(*SnowflakeNode).Generate(), nil
	}
	node, err := NewSnowflakeNode(workerId)
	if err != nil {
		return 0, err
	}
	if actual, loaded := snowflakeNodeMap.LoadOrStore(workerId, node); loaded {
		node = actual.(*SnowflakeNode)
	}

	return node.Generate(), nil
}

func GenerateSnowflakeID(workerId int64) int64 {
	id, _ := NewSnowflakeID(workerId)
	return id
}
