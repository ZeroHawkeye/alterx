package alterx

// ClusterBomb 实现了N阶变量组合爆破算法，用可变长度数组/值
// 该算法用于生成所有可能的变量组合
func ClusterBomb(payloads *IndexMap, callback func(varMap map[string]interface{}), Vector []string) {
	// 算法目标：通过构建向量减少任意值的数量

	// 算法步骤
	// 步骤1) 初始化/输入一个IndexMap(这里是：payloads)
	// indexMap实际上是一个将所有键索引到另一个映射中的映射

	// 步骤2) Vector是长度为n的数组，使得n = len(payloads)
	// payloads(IndexMap)中的每个值包含一个数组
	// 例如：payloads["word"] = []string{"api","dev","cloud"}

	// 步骤3) Vector的初始长度为0。通过递归
	// 我们构建一个包含payloads[N]所有可能值的Vector，其中N = 0 < len(payloads)

	// 步骤4) 在递归结束时，len(Vector) == len(payloads).Cap() - 1
	// 这意味着Vn = {r0,r1,...,rn}，只有rn缺失
	// 在这种情况下，遍历rn的所有可能值，即payload.GetNth(n)
	if len(Vector) == payloads.Cap()-1 {
		// 向量末端
		vectorMap := map[string]interface{}{}
		for k, v := range Vector {
			// 用所有可用向量构建map[变量]=值
			vectorMap[payloads.KeyAtNth(k)] = v
		}
		// 映射中缺少一个元素，即最后一个元素
		index := len(Vector)
		for _, elem := range payloads.GetNth(index) {
			vectorMap[payloads.KeyAtNth(index)] = elem
			callback(vectorMap)
		}
		return
	}

	// 步骤5) 如果向量尚未填充到payload.Cap()-1
	// 遍历第r个变量的有效载荷并使用递归执行它们
	// 如果Vector为空或在第1个索引处固定，则遍历第x个位置
	index := len(Vector)
	for _, v := range payloads.GetNth(index) {
		var tmp []string
		if len(Vector) > 0 {
			tmp = append(tmp, Vector...)
		}
		tmp = append(tmp, v)
		ClusterBomb(payloads, callback, tmp) // 递归调用
	}
}

// IndexMap 是一个特殊的映射结构，允许通过索引访问键和值
type IndexMap struct {
	values  map[string][]string // 存储变量名到变量值列表的映射
	indexes map[int]string      // 存储索引到变量名的映射
}

// GetNth 返回第n个位置的值列表
func (o *IndexMap) GetNth(n int) []string {
	return o.values[o.indexes[n]]
}

// Cap 返回映射中的元素数量
func (o *IndexMap) Cap() int {
	return len(o.values)
}

// KeyAtNth 返回第n个位置的键
func (o *IndexMap) KeyAtNth(n int) string {
	return o.indexes[n]
}

// NewIndexMap 返回一个类型，使得映射的元素可以通过固定索引检索
// 这种实现使得可以通过数字索引访问映射元素，便于处理
func NewIndexMap(values map[string][]string) *IndexMap {
	i := &IndexMap{
		values: values,
	}
	indexes := map[int]string{}
	counter := 0
	for k := range values {
		indexes[counter] = k
		counter++
	}
	i.indexes = indexes
	return i
}
