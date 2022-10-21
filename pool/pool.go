package pool

//var (
//	errorPoolClosed  = fmt.Errorf("conn pool has closed")
//	errorPoolFull    = fmt.Errorf("conn pool is full")
//	errorPoolTimeout = fmt.Errorf("conn pool wait timeout")
//)
//
//type connReuseStrategy int
//
//const (
//	CachedOrNewConn connReuseStrategy = iota
//	OnlyCached
//)
//
//type Pool interface {
//}
//
//type ConnOptions struct {
//	MaxConn       int
//	MaxIdle       int
//	WaitTimeout   time.Duration
//	MaxLifetime   time.Duration
//	ReuseStrategy connReuseStrategy
//}
//
//type conn struct {
//	lastUsedAt time.Time
//	connection io.Closer
//}
//
//type nextConnIndex struct {
//	index int
//}
//
//type token struct {
//	nextConnIndex
//	conn        io.Closer
//	createdAt   time.Time     // 创建时间
//	maxLifetime time.Duration // 最大存活时间
//}
//
//type PoolImpl struct {
//	// control params
//	mu            sync.Mutex
//	conns         []int
//	factory       func() (io.Closer, error)
//	closed        bool
//	freeConn      int
//	waitQueue     map[int]chan token
//	nextConnIndex nextConnIndex
//	freeConns     map[int]token
//	openCnt       int
//	waitCnt       int
//
//	// setting params
//	maxConn     int
//	maxIdle     int
//	maxLifetime time.Duration
//	waitTimeout time.Duration
//	strategy    connReuseStrategy
//}
//
//func NewPool(ctx context.Context, opts *ConnOptions) Pool {
//	return &PoolImpl{
//		maxConn:   opts.MaxConn,
//		maxIdle:   opts.MaxIdle,
//		openCnt:   0,
//		conns:     make([]int, 0),
//		waitQueue: make(map[int]chan token),
//		freeConns: make(map[int]token),
//	}
//}
//
//func (p *PoolImpl) Get(ctx context.Context) (io.Closer, error) {
//	p.mu.Lock()
//
//	// check pool closed
//	if p.closed {
//		p.mu.Unlock()
//		return nil, errorPoolClosed
//	}
//
//	// check context
//	select {
//	case <-ctx.Done():
//		p.mu.Unlock()
//		return nil, ctx.Err()
//	default:
//	}
//
//	if len(p.freeConns) > 0 {
//		var popToken token
//		var popReqKey int
//
//		for popReqKey, popToken = range p.freeConns {
//			break
//		}
//
//		delete(p.freeConns, popReqKey)
//		p.mu.Unlock()
//		return popToken.conn, nil
//	}
//
//	if p.openCnt >= p.maxConn {
//		nextConnIndex := p.getNextConnIndex()
//
//		req := make(chan token, 1)
//		p.waitQueue[nextConnIndex] = req
//		p.waitCnt++
//		p.mu.Unlock()
//
//		select {
//		case <-time.After(p.waitTimeout):
//			return nil, errorPoolTimeout
//		case t, ok := <-req:
//			if !ok {
//				return nil, errorPoolFull // 这里是什么错误?
//			}
//			return t.conn, nil
//		case <-ctx.Done():
//			return nil, ctx.Err()
//		}
//	}
//
//	p.openCnt++
//	p.mu.Unlock()
//
//	t := token{nextConnIndex: p.nextConnIndex}
//
//}
//
//func (p *PoolImpl) getNextConnIndex() int {
//	currIndex := p.nextConnIndex.index
//	p.nextConnIndex.index = currIndex + 1
//	return p.nextConnIndex.index
//}
