package port

type (
	// Mapper converts values in both directions at an application boundary.
	Mapper[I, O any] interface {
		To[I, O]
		From[O, I]
	}
	// To converts one representation into another.
	To[I, O any] interface {
		To(I) O
	}
	// From converts a source representation into the target representation.
	From[I, O any] interface {
		From(I) O
	}
)
