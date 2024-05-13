package consumer

// receives and processes events using fetcher and processor
type Consumer interface {
	Start() error
}