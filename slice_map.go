package slice_map

const (
	_BASE_NO_SHRINK_SIZE = 1024
)

//该接口代表要映射的元素结构体对象
type LMapObj interface {
	LMapId() int	//返回映射对象的ID
}

//缓存对象LMAP
type LMap struct {
	objSlice []LMapObj	//映射对象组成的数组切片
	idxMap   map[int]int	//存储对象索引的MAP
	maxIdx   int	//最大的索引值
}

//缓存的构造函数
func NewLMap() *LMap {
	return &LMap{
		//初始化元素切片和MAP
		objSlice: make([]LMapObj, 0),
		idxMap:   make(map[int]int),
		maxIdx:   0,
	}
}

func (lmap *LMap) Len() int {
	return lmap.maxIdx
}

//通过缓存key返回缓存对象
func (lmap *LMap) Get(id int) LMapObj {
	if idx, present := lmap.idxMap[id]; present {
		return lmap.objSlice[idx]
	}
	return nil
}

//删除指定的元素
func (lmap *LMap) Del(id int) {
	if curIdx, present := lmap.idxMap[id]; present {
		delete(lmap.idxMap, id)
		//如果是最后一个元素,就直接置空
		if curIdx == lmap.maxIdx-1 {
			lmap.objSlice[curIdx] = nil
			lmap.maxIdx--
		} else {
			//否则就将最后的元素复制到要删除的位置
			lmap.maxIdx--
			//从数组中获得真正的缓存对象
			obj := lmap.objSlice[lmap.maxIdx]
			//将缓存对象赋值给指定元素
			lmap.objSlice[curIdx] = obj
			//将原数据id和最新的缓存对象下标建立关联
			lmap.idxMap[obj.LMapId()] = curIdx
			lmap.objSlice[lmap.maxIdx] = nil
		}
	} else {
		// errors.New("Id not found.")
		return
	}

	// Shrink to prevent slice increasing with no limit.
	//计算空置的元素个数
	fvalue := len(lmap.objSlice) - lmap.maxIdx
	//如果最大索引超过界定,空置数超过10%,就执行收缩操作
	if lmap.maxIdx > _BASE_NO_SHRINK_SIZE && fvalue > 0 &&
		(float32(fvalue)/float32(lmap.maxIdx) > 0.1) {
		lmap.shrink()
	}
}

func (lmap *LMap) Add(obj LMapObj) {
	//获得缓存数据的ID
	id := obj.LMapId()
	//缓存数据的ID在映射中对应的下标如果存在值,就返回缓存数组中存储的缓存对象
	if idx, exist := lmap.idxMap[id]; exist {
		lmap.objSlice[idx] = obj
		return
	}

	//否则就将最大索引存入映射中
	lmap.idxMap[id] = lmap.maxIdx
	//如果最大索引在数组长度内,就直接赋值
	if lmap.maxIdx < len(lmap.objSlice) {
		lmap.objSlice[lmap.maxIdx] = obj
	} else {
		//否则就在数组末尾追加
		lmap.objSlice = append(lmap.objSlice, obj)
	}
	//最大索引自增,以备后用
	lmap.maxIdx++
}

//收缩操作
func (lmap *LMap) shrink() {
	//如果最大索引在数组长度内,说明数组正常,即可按最大索引剪辑数组
	if lmap.maxIdx < len(lmap.objSlice) {
		lmap.objSlice = lmap.objSlice[0:lmap.maxIdx]
	}
}
//公开调用的方法
func (lmap *LMap) Shrink() {
	lmap.shrink()
}

//快速迭代方法
// Iterting map, as you must not delete any element in user-specified function.
func (lmap *LMap) FastIter(f func(LMapObj)) {
	//遍历缓存数组的索引和值
	for _, c := range lmap.objSlice {
		if c == nil {
			break
		}
		//将缓存数据传入闭包执行
		f(c)
	}
}

// Iterating map with the ability of deleting element.
// But some element would not be travelled, if you deleted keys.
func (lmap *LMap) Iter(f func(LMapObj)) {
	i := 0
	for _, c := range lmap.objSlice {
		if c == nil {
			break
		}

		f(c)
		//遍历缓存数组
		if i < len(lmap.objSlice) {
			//获取缓存的元素对象
			newC := lmap.objSlice[i]
			//如果元素对象不为空,且和旧的缓存对象不一致,就对新的元素对象进行闭包处理
			for newC != nil && newC.LMapId() != c.LMapId() {
				f(newC)
				//执行完后,将新的元素对象赋值给旧的缓存对象,以便能退出循环
				c = newC
				if i < len(lmap.objSlice) {
					//将闭包处理后的元素对象再回写给新的元素对象
					newC = lmap.objSlice[i]
				}
			}
		}
		i++
	}
}
