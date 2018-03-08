package como

var waitCloseChan chan struct{}

// 等待como关闭
func WaitClose() {
	<-waitCloseChan
}

// como关闭
func Close() {
	select {
	case waitCloseChan <- struct{}{}:
	default:
	}
}

// 初始化como
func init() {
	// 创建等待关闭通道
	waitCloseChan = make(chan struct{}, 1)
	pactInit()
}
