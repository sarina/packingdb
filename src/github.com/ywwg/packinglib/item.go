package packinglib

import (
	"fmt"
	"math"
)

const NoUnits = "nounits"

type Item interface {
	// Satisfies returns true if the item belongs in the given context
	Satisfies(*Context) bool

	// Pack tells the item to pack itself given the context and returns the packed item
	Pack(*Trip) Item

	// Count returns the number of this item that got packed
	Count() float64

	// String prints out a string representation of the packed item(s)
	String() string
}

// BasicItem is the simplest item -- just prerequisites and no count, like "tent"
type BasicItem struct {
	// Name of the item.
	Name string

	// count is the number of this thing that got packed.
	count float64

	// Prerequisites is a set of all properties that the context must have for this item to appear.
	Prerequisites PropertySet
}

func NewBasicItem(name string, allow, disallow []string) *BasicItem {
	return &BasicItem{
		Name:          name,
		Prerequisites: buildPropertySet(allow, disallow),
	}
}

// Satisfies returns true if the context satisfies the item's requirements.
func (i *BasicItem) Satisfies(c *Context) bool {
	// Any property satisfies (OR)
	if len(i.Prerequisites) == 0 {
		return true
	}
	found := false
	for p := range i.Prerequisites {
		// Any item that has a disallowing prerequisite immediately dissatisfies.
		if allow, ok := c.Properties[p]; ok {
			if !allow {
				return false
			}
			found = true
		}
	}
	return found
}

func (i *BasicItem) Pack(t *Trip) Item {
	p := &BasicItem{}
	*p = *i
	if p.Satisfies(t.C) {
		p.count = 1.0
	}
	return p
}

func (i *BasicItem) Count() float64 {
	return i.count
}

func (i *BasicItem) String() string {
	return i.Name
}

type TemperatureItem struct {
	BasicItem

	// TemperatureMin is the anticipated minimum temperature.
	TemperatureMin int

	// TemperatureMax is the anticipated maximum temperature.
	TemperatureMax int
}

func NewTemperatureItem(name string, min, max int, allow, disallow []string) *TemperatureItem {
	return &TemperatureItem{
		BasicItem:      *NewBasicItem(name, allow, disallow),
		TemperatureMin: min,
		TemperatureMax: max,
	}
}

func (i *TemperatureItem) Satisfies(c *Context) bool {
	if i.TemperatureMax < c.TemperatureMin {
		return false
	}
	if i.TemperatureMin > c.TemperatureMax {
		return false
	}

	return i.BasicItem.Satisfies(c)
}

func (i *TemperatureItem) Pack(t *Trip) Item {
	p := &TemperatureItem{}
	*p = *i
	if p.Satisfies(t.C) {
		p.count = 1.0
	}
	return p
}

type ConsumableItem struct {
	BasicItem

	// DailyRate is how much the thing gets used per day.
	DailyRate float64

	// What units the rate is in.  Use NoUnits for things without "of" qualifiers. ("1 car")
	Units string

	// Prerequisites is a set of all properties that the context must have for this item to appear.
	Prerequisites map[Property]bool
}

func NewConsumableItem(name string, rate float64, units string, allow, disallow []string) *ConsumableItem {
	return &ConsumableItem{
		BasicItem: *NewBasicItem(name, allow, disallow),
		DailyRate: rate,
		Units:     units,
	}
}

func (i *ConsumableItem) Pack(t *Trip) Item {
	p := &ConsumableItem{}
	*p = *i
	if p.Satisfies(t.C) {
		p.count = math.Ceil(i.DailyRate * float64(t.Days))
	}
	return p
}

func (i *ConsumableItem) String() string {
	if i.Units == NoUnits {
		if i.count == float64(int(i.count)) {
			return fmt.Sprintf("%d %s", int(i.count), i.Name)
		} else {
			return fmt.Sprintf("%.1f %s", i.count, i.Name)
		}
	} else {
		if i.count == float64(int(i.count)) {
			return fmt.Sprintf("%d %s of %s", int(i.count), i.Units, i.Name)
		} else {
			return fmt.Sprintf("%.1f %s of %s", i.count, i.Units, i.Name)
		}
	}
}

type CustomConsumableItem struct {
	ConsumableItem

	// DailyRate is how much the thing gets used per day.
	RateFunc func(days int) float64
}

func NewCustomConsumableItem(name string, rateFunc func(days int) float64, units string, allow, disallow []string) *CustomConsumableItem {
	return &CustomConsumableItem{
		ConsumableItem: *NewConsumableItem(name, 0, units, allow, disallow),
		RateFunc:       rateFunc,
	}
}

func (i *CustomConsumableItem) Pack(t *Trip) Item {
	p := &CustomConsumableItem{}
	*p = *i
	if p.Satisfies(t.C) {
		p.count = i.RateFunc(t.Days)
	}
	return p
}

type ConsumableTemperatureItem struct {
	ConsumableItem
	TemperatureItem
}

func NewConsumableTemperatureItem(name string, rate float64, units string, min, max int, allow, disallow []string) *ConsumableTemperatureItem {
	return &ConsumableTemperatureItem{
		ConsumableItem:  *NewConsumableItem(name, rate, units, allow, disallow),
		TemperatureItem: *NewTemperatureItem(name, min, max, allow, disallow),
	}
}

func (i *ConsumableTemperatureItem) Satisfies(c *Context) bool {
	if !i.TemperatureItem.Satisfies(c) {
		return false
	}

	return i.ConsumableItem.Satisfies(c)
}

func (i *ConsumableTemperatureItem) Pack(t *Trip) Item {
	p := &ConsumableTemperatureItem{}
	*p = *i
	if p.Satisfies(t.C) {
		p.ConsumableItem.count = math.Ceil(i.DailyRate * float64(t.Days))
	}
	return p
}

func (i *ConsumableTemperatureItem) Count() float64 {
	return i.ConsumableItem.count
}

func (i *ConsumableTemperatureItem) String() string {
	return i.ConsumableItem.String()
}
