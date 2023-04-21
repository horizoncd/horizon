package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type keylogger string

// Field is an alias for the field structure in the underlying log frame.
type Field = zapcore.Field

// Level is an alias for the level structure in the underlying log frame.
type Level = zapcore.Level

// Defines common log fields.
const (
	KeyRequestID   keylogger = "requestID"
	KeyUsername    keylogger = "username"
	KeyWatcherName keylogger = "watcher"
)

// Define the basic log level.
// Alias for zap level type const.
var (
	// DebugLevel logs are typically voluminous, and are usually disabled in production.
	DebugLevel = zapcore.DebugLevel
	// InfoLevel is the default logging priority.
	InfoLevel = zapcore.InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual human review.
	WarnLevel = zapcore.WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly, it shouldn't generate any error-level logs.
	ErrorLevel = zapcore.ErrorLevel
	// PanicLevel logs a message, then panics.
	PanicLevel = zapcore.PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel = zapcore.FatalLevel
)

// Alias for zap type functions.
var (
	Any         = zap.Any         // accepts a field key and an arbitrary value
	Array       = zap.Array       // accepts a field key and an ArrayMarshaler
	Bool        = zap.Bool        // accepts a field key and a bool
	Boolp       = zap.Boolp       // accepts a field key and a *bool
	Bools       = zap.Bools       // accepts a field key and a []bool
	ByteString  = zap.ByteString  // accepts a field key and a ByteString
	ByteStrings = zap.ByteStrings // accepts a field key and a []ByteString
	Binary      = zap.Binary      // accepts a field key and []byte
	Complex128  = zap.Complex128  // accepts a field key and a complex128
	Complex128p = zap.Complex128p // accepts a field key and a *complex128
	Complex128s = zap.Complex128s // accepts a field key and a []complex128
	Complex64   = zap.Complex64   // accepts a field key and a complex64
	Complex64p  = zap.Complex64p  // accepts a field key and a *complex64
	Complex64s  = zap.Complex64s  // accepts a field key and a []complex64
	Duration    = zap.Duration    // accepts a field key and a time.Duration
	Durationp   = zap.Durationp   // accepts a field key and a *time.Duration
	Durations   = zap.Durations   // accepts a field key and a []time.Duration
	Err         = zap.Error       // accepts a field key and an error
	Errors      = zap.Errors      // accepts a field key and a []error
	Float32     = zap.Float32     // accepts a field key and a float32
	Float32p    = zap.Float32p    // accepts a field key and a *float32
	Float32s    = zap.Float32s    // accepts a field key and a []float32
	Float64     = zap.Float64     // accepts a field key and a float64
	Float64p    = zap.Float64p    // accepts a field key and a *float64
	Float64s    = zap.Float64s    // accepts a field key and a []float64
	Int         = zap.Int         // accepts a field key and an int
	Intp        = zap.Intp        // accepts a field key and a *int
	Ints        = zap.Ints        // accepts a field key and a []int
	Int64       = zap.Int64       // accepts a field key and an int64
	Int64p      = zap.Int64p      // accepts a field key and a *int64
	Int64s      = zap.Int64s      // accepts a field key and a []int64
	Int32       = zap.Int32       // accepts a field key and an int32
	Int32p      = zap.Int32p      // accepts a field key and a *int32
	Int32s      = zap.Int32s      // accepts a field key and a []int32
	Int16       = zap.Int16       // accepts a field key and an int16
	Int16p      = zap.Int16p      // accepts a field key and a *int16
	Int16s      = zap.Int16s      // accepts a field key and a []int16
	Int8        = zap.Int8        // accepts a field key and an int8
	Int8p       = zap.Int8p       // accepts a field key and a *int8
	Int8s       = zap.Int8s       // accepts a field key and a []int8
	Namespace   = zap.Namespace   // accepts a field key and returns a new logger that is a child of the current logger and adds a namespace to the field key
	Object      = zap.Object      // accepts a field key and an ObjectMarshaler
	Reflect     = zap.Reflect     // accepts a field key and an arbitrary value, and reflects on the value to create more fields
	Skip        = zap.Skip        // accepts no arguments and is a no-op
	String      = zap.String      // accepts a field key and a string
	Stringp     = zap.Stringp     // accepts a field key and a *string
	Strings     = zap.Strings     // accepts a field key and a []string
	Stringer    = zap.Stringer    // accepts a field key and a fmt.Stringer
	Time        = zap.Time        // accepts a field key and a time.Time
	Timep       = zap.Timep       // accepts a field key and a *time.Time
	Times       = zap.Times       // accepts a field key and a []time.Time
	Uint        = zap.Uint        // accepts a field key and a uint
	Uintp       = zap.Uintp       // accepts a field key and a *uint
	Uints       = zap.Uints       // accepts a field key and a []uint
	Uint64      = zap.Uint64      // accepts a field key and a uint64
	Uint64p     = zap.Uint64p     // accepts a field key and a *uint64
	Uint64s     = zap.Uint64s     // accepts a field key and a []uint64
	Uint32      = zap.Uint32      // accepts a field key and a uint32
	Uint32p     = zap.Uint32p     // accepts a field key and a *uint32
	Uint32s     = zap.Uint32s     // accepts a field key and a []uint32
	Uint16      = zap.Uint16      // accepts a field key and a uint16
	Uint16p     = zap.Uint16p     // accepts a field key and a *uint16
	Uint16s     = zap.Uint16s     // accepts a field key and a []uint16
	Uint8       = zap.Uint8       // accepts a field key and a uint8
	Uint8p      = zap.Uint8p      // accepts a field key and a *uint8
	Uint8s      = zap.Uint8s      // accepts a field key and a []uint8
	Uintptr     = zap.Uintptr     // accepts a field key and a uintptr
	Uintptrp    = zap.Uintptrp    // accepts a field key and a *uintptr
	Uintptrs    = zap.Uintptrs    // accepts a field key and a []uintptr
)
