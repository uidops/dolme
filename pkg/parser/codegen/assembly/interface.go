package assembly

// Assembly interface defines methods for generating and building assembly code.
type Assembly interface {
	Generate() error
	GetCode() string
	Build() error
}
