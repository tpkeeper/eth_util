package notify

type Notifier interface {
	Notify(chatId string, str string) error
}
