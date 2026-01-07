package plc

import (
	"fmt"
	"sync"
	"time"
)

// TagType represents the data type of a tag
type TagType int

const (
	TagTypeFloat64 TagType = iota
	TagTypeInt32
	TagTypeBool
	TagTypeString
)

// Tag represents a PLC tag (variable)
type Tag struct {
	Name        string      // Tag name (same as sensor name)
	Type        TagType     // Data type
	Value       interface{} // Current value
	Address     string      // PLC address (%DF100, %MW0, etc)
	Description string      // Tag description
	Quality     bool        // Data quality (good/bad)
	Timestamp   time.Time   // Last update timestamp
	mu          sync.RWMutex
}

// NewTag creates a new tag
func NewTag(name, address, description string, tagType TagType) *Tag {
	var defaultValue interface{}
	switch tagType {
	case TagTypeFloat64:
		defaultValue = 0.0
	case TagTypeInt32:
		defaultValue = int32(0)
	case TagTypeBool:
		defaultValue = false
	case TagTypeString:
		defaultValue = ""
	}

	return &Tag{
		Name:        name,
		Type:        tagType,
		Value:       defaultValue,
		Address:     address,
		Description: description,
		Quality:     true,
		Timestamp:   time.Now(),
	}
}

// GetValue returns the tag value (thread-safe)
func (t *Tag) GetValue() interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Value
}

// GetFloat64 returns the tag value as float64
func (t *Tag) GetFloat64() (float64, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	switch v := t.Value.(type) {
	case float64:
		return v, nil
	case int32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0, fmt.Errorf("tag %s: cannot convert %T to float64", t.Name, t.Value)
	}
}

// GetInt32 returns the tag value as int32
func (t *Tag) GetInt32() (int32, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	switch v := t.Value.(type) {
	case int32:
		return v, nil
	case int:
		return int32(v), nil
	case float64:
		return int32(v), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("tag %s: cannot convert %T to int32", t.Name, t.Value)
	}
}

// GetBool returns the tag value as bool
func (t *Tag) GetBool() (bool, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	switch v := t.Value.(type) {
	case bool:
		return v, nil
	case int32:
		return v != 0, nil
	case int:
		return v != 0, nil
	case float64:
		return v != 0, nil
	default:
		return false, fmt.Errorf("tag %s: cannot convert %T to bool", t.Name, t.Value)
	}
}

// SetValue sets the tag value (thread-safe)
func (t *Tag) SetValue(value interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Type validation
	switch t.Type {
	case TagTypeFloat64:
		switch v := value.(type) {
		case float64:
			t.Value = v
		case int:
			t.Value = float64(v)
		case int32:
			t.Value = float64(v)
		default:
			return fmt.Errorf("tag %s: invalid type %T for Float64 tag", t.Name, value)
		}
	case TagTypeInt32:
		switch v := value.(type) {
		case int32:
			t.Value = v
		case int:
			t.Value = int32(v)
		case float64:
			t.Value = int32(v)
		default:
			return fmt.Errorf("tag %s: invalid type %T for Int32 tag", t.Name, value)
		}
	case TagTypeBool:
		switch v := value.(type) {
		case bool:
			t.Value = v
		case int:
			t.Value = v != 0
		case int32:
			t.Value = v != 0
		case float64:
			t.Value = v != 0
		default:
			return fmt.Errorf("tag %s: invalid type %T for Bool tag", t.Name, value)
		}
	case TagTypeString:
		t.Value = fmt.Sprintf("%v", value)
	}

	t.Timestamp = time.Now()
	return nil
}

// SetQuality sets the data quality
func (t *Tag) SetQuality(quality bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Quality = quality
}

// GetQuality returns the data quality
func (t *Tag) GetQuality() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Quality
}

// GetTimestamp returns the last update timestamp
func (t *Tag) GetTimestamp() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Timestamp
}

// TagManager manages all PLC tags
type TagManager struct {
	tags map[string]*Tag
	mu   sync.RWMutex
}

// NewTagManager creates a new tag manager
func NewTagManager() *TagManager {
	return &TagManager{
		tags: make(map[string]*Tag),
	}
}

// AddTag adds a new tag
func (tm *TagManager) AddTag(tag *Tag) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.tags[tag.Name]; exists {
		return fmt.Errorf("tag '%s' already exists", tag.Name)
	}

	tm.tags[tag.Name] = tag
	return nil
}

// GetTag retrieves a tag by name
func (tm *TagManager) GetTag(name string) (*Tag, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tag, exists := tm.tags[name]
	if !exists {
		return nil, fmt.Errorf("tag '%s' not found", name)
	}

	return tag, nil
}

// GetAllTags returns all tags
func (tm *TagManager) GetAllTags() []*Tag {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tags := make([]*Tag, 0, len(tm.tags))
	for _, tag := range tm.tags {
		tags = append(tags, tag)
	}
	return tags
}

// TagExists checks if a tag exists
func (tm *TagManager) TagExists(name string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	_, exists := tm.tags[name]
	return exists
}

// GetTagValue gets tag value by name
func (tm *TagManager) GetTagValue(name string) (interface{}, error) {
	tag, err := tm.GetTag(name)
	if err != nil {
		return nil, err
	}
	return tag.GetValue(), nil
}

// SetTagValue sets tag value by name
func (tm *TagManager) SetTagValue(name string, value interface{}) error {
	tag, err := tm.GetTag(name)
	if err != nil {
		return err
	}
	return tag.SetValue(value)
}

// GetTagCount returns the number of tags
func (tm *TagManager) GetTagCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.tags)
}
