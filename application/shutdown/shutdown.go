package shutdown

var stopBridge = false
var tendermintStopped = false
var ethStopped = false
var kafkaConsumerClosed = false

func GetBridgeStopSignal() bool {
	return stopBridge
}

func SetBridgeStopSignal(value bool) {
	stopBridge = value
}

func GetTMStopped() bool {
	return tendermintStopped
}

func SetTMStopped(value bool) {
	tendermintStopped = value
}

func GetETHStopped() bool {
	return ethStopped
}

func SetETHStopped(value bool) {
	ethStopped = value
}

func GetKafkaConsumerClosed() bool {
	return kafkaConsumerClosed
}

func SetKafkaConsumerClosed(value bool) {
	kafkaConsumerClosed = value
}
