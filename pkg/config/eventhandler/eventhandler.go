package eventhandler

type Config struct {
	// how many events to process in one batch
	BatchEventsCount uint `yaml:"batchEventsCount"`
	// seconds to save after each cursor save
	CursorSaveInterval uint `yaml:"cursorSaveInterval"`
	// seconds to wait when there is no events
	IdleWaitInterval uint `yaml:"idleWaitInterval"`
}
