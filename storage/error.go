package storage

type StorageError struct {
	msg string
}

func (sr StorageError) Error() string {
	return sr.msg
}

func newStorageError(msg string) error {
	return StorageError{
		msg: msg,
	}
}

var (
	errInvalidStorageType = newStorageError("err invalid storage type")
	errInvalidConfig      = newStorageError("err invalid config")
)
