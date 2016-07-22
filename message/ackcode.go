package message

type AckCode byte

const (
	ConnectionAccepted AckCode = iota

	ErrServerUnavailable

	StartSuccess

	ErrStart

	StopSuccess
	
	ErrStop

	HeartBeatAck

)

func (this AckCode) Value() byte {
	return byte(this)
}

func (this AckCode) Error() string {
	switch this {
	case 0:
		return "Connection accepted"
	case 1:
		return "Connection Refused, Server unavailable"

	case 2:
		return "Start process success"

	case 3:
		return "Fail to start process"

	case 4:
                return "Stop process success"

        case 5:
                return "Fail to stop process"

	case 6:
		return "Heart beat accepted"

	}

	return "Unknown error"
}
