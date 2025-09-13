package interpreter

// Frame represents a function call frame.
type Frame struct {
	FuncName   string        // function name for this frame
	IP         int           // instruction pointer for this frame (index into PB)
	Locals     map[int]Value // local variables (address -> value)
	ReturnToIP int           // IP in caller to continue after return placement
	RetTemp    int           // address where return value should be stored in caller
}
