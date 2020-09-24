package logger

// options keep settings for loggers/sweepers
// Indexed by the logger unique name
type options struct {
	values map[string][]OptionItem
}

// Option represents settings of the logger
type Option struct {
	// Apply logger option
	Apply func(op *options)
}

// BackendOption creates option for the specified backend.
func BackendOption(name string, level string, settings map[string]interface{}) Option {
	return Option{func(op *options) {
		vals := make([]OptionItem, 0)
		vals = append(vals, OptionItem{"level", level})

		// Append extra settings if existing
		if len(settings) > 0 {
			for k, v := range settings {
				vals = append(vals, OptionItem{k, v})
			}
		}

		// Append with overriding way
		op.values[name] = vals
	}}
}

// SweeperOption creates option for the sweeper.
func SweeperOption(name string, duration int, settings map[string]interface{}) Option {
	return Option{func(op *options) {
		vals := make([]OptionItem, 0)
		vals = append(vals, OptionItem{"duration", duration})

		// Append settings if existing
		if len(settings) > 0 {
			for k, v := range settings {
				vals = append(vals, OptionItem{k, v})
			}
		}

		// Append with overriding way
		op.values[name] = vals
	}}
}

// GetterOption creates option for the getter.
func GetterOption(name string, settings map[string]interface{}) Option {
	return Option{func(op *options) {
		vals := make([]OptionItem, 0)
		// Append settings if existing
		if len(settings) > 0 {
			for k, v := range settings {
				vals = append(vals, OptionItem{k, v})
			}
		}

		// Append with overriding way
		op.values[name] = vals
	}}
}

// OptionItem is a simple wrapper of property and value
type OptionItem struct {
	field string
	val   interface{}
}

// Field returns name of the option
func (o *OptionItem) Field() string {
	return o.field
}

// Int returns the integer value of option
func (o *OptionItem) Int() int {
	if o.val == nil {
		return 0
	}

	return o.val.(int)
}

// String returns the string value of option
func (o *OptionItem) String() string {
	if o.val == nil {
		return ""
	}

	return o.val.(string)
}

// Raw returns the raw value
func (o *OptionItem) Raw() interface{} {
	return o.val
}
