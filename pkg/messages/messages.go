package messages

type ChatCompletionMsg struct{}

type SystemMessageSavedMsg struct {
	Message string
}

type SystemPopupStatusMsg struct {
	IsOpen bool
}

type GoBackMsg struct{}

type PushViewMsg struct {
	View int
}
