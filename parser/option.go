package parser

// NullStyle .
type NullStyle int

const (
	// NullDisable .
	NullDisable NullStyle = iota

	// NullInSQL .
	NullInSQL

	// NullInPointer .
	NullInPointer
)

// Option .
type Option func(*options)

// options .
type options struct {
	Charset        string    `json:"-"`
	Collation      string    `json:"-"`
	JSONTag        bool      `json:"-"`
	TablePrefix    string    `json:"-"`
	ColumnPrefix   string    `json:"-"`
	NoNullType     bool      `json:"-"`
	NullStyle      NullStyle `json:"-"`
	Package        string    `json:"-"`
	GormType       bool      `json:"-"`
	ForceTableName bool      `json:"-"`
}

// defaultOptions .
var defaultOptions = options{
	NullStyle: NullInSQL,
	Package:   "model",
}

// WithCharset .
func WithCharset(charset string) Option {
	return func(o *options) {
		o.Charset = charset
	}
}

// WithCollation .
func WithCollation(collation string) Option {
	return func(o *options) {
		o.Collation = collation
	}
}

// WithTablePrefix .
func WithTablePrefix(p string) Option {
	return func(o *options) {
		o.TablePrefix = p
	}
}

// WithColumnPrefix .
func WithColumnPrefix(p string) Option {
	return func(o *options) {
		o.ColumnPrefix = p
	}
}

// WithJSONTag .
func WithJSONTag() Option {
	return func(o *options) {
		o.JSONTag = true
	}
}

// WithNoNullType .
func WithNoNullType() Option {
	return func(o *options) {
		o.NoNullType = true
	}
}

// WithNullStyle .
func WithNullStyle(s NullStyle) Option {
	return func(o *options) {
		o.NullStyle = s
	}
}

// WithPackage .
func WithPackage(pkg string) Option {
	return func(o *options) {
		o.Package = pkg
	}
}

// WithGormType will write type in gorm tag
func WithGormType() Option {
	return func(o *options) {
		o.GormType = true
	}
}

// WithForceTableName .
func WithForceTableName() Option {
	return func(o *options) {
		o.ForceTableName = true
	}
}

// parseOption .
func parseOption(options []Option) options {
	o := defaultOptions
	for _, f := range options {
		f(&o)
	}
	if o.NoNullType {
		o.NullStyle = NullDisable
	}
	return o
}
