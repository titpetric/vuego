package vuego

// SlotScope holds all slot contents indexed by name for a component instance.
type SlotScope struct {
	// Slots maps slot names to their content.
	// Empty string key is the default slot.
	Slots map[string]*SlotContent
}

// NewSlotScope creates a new SlotScope for a component.
func NewSlotScope() *SlotScope {
	return &SlotScope{
		Slots: make(map[string]*SlotContent),
	}
}

// GetSlot retrieves slot content by name.
// Returns nil if the slot doesn't exist.
func (ss *SlotScope) GetSlot(name string) *SlotContent {
	return ss.Slots[name]
}

// SetSlot sets the content for a named slot.
func (ss *SlotScope) SetSlot(name string, content *SlotContent) {
	ss.Slots[name] = content
}
