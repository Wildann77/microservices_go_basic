package response

// Meta contains pagination metadata for list responses
type Meta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// NewMeta creates a new Meta instance
func NewMeta(total, limit, offset int) Meta {
	return Meta{
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
}

// WithTotal returns a new Meta with updated total
func (m Meta) WithTotal(total int) Meta {
	m.Total = total
	return m
}

// WithLimit returns a new Meta with updated limit
func (m Meta) WithLimit(limit int) Meta {
	m.Limit = limit
	return m
}

// WithOffset returns a new Meta with updated offset
func (m Meta) WithOffset(offset int) Meta {
	m.Offset = offset
	return m
}
