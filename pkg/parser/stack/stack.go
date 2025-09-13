package stack

type Stack struct {
	a []string
	l int
}

// NewStack creates a new stack instance
func NewStack(elm ...string) *Stack {
	stack := Stack{
		a: make([]string, 0),
		l: 0,
	}

	for _, e := range elm {
		stack.l++
		stack.a = append(stack.a, e)
	}

	return &stack
}

// Push adds an element to the top of the stack
func (s *Stack) Push(elm string) {
	s.l++
	s.a = append(s.a, elm)
}

// Pop removes and returns the top element of the stack
func (s *Stack) Pop() string {
	if s.l < 1 {
		return ""
	}

	s.l--
	elm := s.a[s.l]
	s.a = s.a[:s.l]

	return elm
}

// Peek returns the top element of the stack without removing it
func (s *Stack) Peek() string {
	if s.l < 1 {
		return ""
	}

	return s.a[s.l-1]
}

// Get the size of the stack
func (s *Stack) Size() int {
	return s.l
}

// Array returns the underlying array of the stack
func (s Stack) Array() []string {
	return s.a
}
