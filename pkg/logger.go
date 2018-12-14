package pkg

type Logger struct {
}

func (logger Logger) Write(data []byte) (count int, error error) {
	return len(data), nil
}
