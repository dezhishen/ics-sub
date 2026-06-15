package subscriptions

import (
	"fmt"
	"sort"
)

var registry = map[string]Provider{}

// Register adds a provider plugin to the in-memory registry.
func Register(p Provider) {
	name := p.Name()
	if name == "" {
		panic("subscriptions provider name cannot be empty")
	}
	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("subscriptions provider already registered: %s", name))
	}
	registry[name] = p
}

// GenerateAll executes all registered plugins and merges the results.
func GenerateAll() ([]Calendar, error) {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)

	all := make([]Calendar, 0)
	for _, name := range names {
		provider := registry[name]
		if provider.Disabled() {
			continue
		}
		items, err := provider.Generate()
		if err != nil {
			return nil, fmt.Errorf("provider %s failed: %w", name, err)
		}
		for i := range items {
			items[i].Provider = name
		}
		all = append(all, items...)
	}

	return all, nil
}
