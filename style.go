package table

type TableStyle struct {
	// DefaultWidth defines the default width of the table.
	DefaultWidth int
	// FitToTerminal defines if the table should fit to the terminal width.
	FitToTerminal bool

	// WrapText defines if the text should be wrapped.
	WrapText bool
	// HideEmpty defines if empty rows should be hidden.
	HideEmpty bool

	// OuterPadding defines the padding around the table.
	OuterPadding int
	// InnerPadding defines the padding between the cells.
	InnerPadding int
}

var defaultTableStyle = &TableStyle{
	DefaultWidth:  80,
	FitToTerminal: true,
	WrapText:      false,
	HideEmpty:     true,
	OuterPadding:  0,
	InnerPadding:  1,
}
