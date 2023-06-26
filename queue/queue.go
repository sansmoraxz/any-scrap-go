package queue

type Queue interface {
	// Close closes the underlying client object.
	Close() error

	// ReadMessage reads a single message from the underlying client object and then calls the callback function
	// passing the message as a parameter.
	// Commit the message if the callback function returns nil
	ReadMessage(callback func([]byte) error) ([]byte, error)

	// WriteMessage writes a single message to the underlying client object and then calls the callback function.
	WriteMessage(msg []byte, callback func([]byte) error) error
}
