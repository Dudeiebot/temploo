package helpers

type (
	messageInterface map[string]interface{}
	messageMap       map[string]string
)

func Message(message string) messageMap {
	return messageMap{
		"message": message,
	}
}

func Response(key string, value interface{}) messageInterface {
	return messageInterface{
		key: value,
	}
}
