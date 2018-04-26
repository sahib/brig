package catfs

import (
	h "github.com/sahib/brig/util/hashlib"
)

type pinCacheEntry struct {
	isExplicit bool
	isPinned   bool
}

type PinCache struct {
	cache map[string]pinCacheEntry
}

func NewPinCache() *PinCache {
	return &PinCache{
		cache: make(map[string]pinCacheEntry),
	}
}

func (pc *PinCache) Remember(content h.Hash, isPinned, isExplicit bool) {
	pc.cache[content.B58String()] = pinCacheEntry{
		isPinned:   isPinned,
		isExplicit: isExplicit,
	}
}

func (pc *PinCache) Is(content h.Hash) (bool, bool, bool) {
	entry, ok := pc.cache[content.B58String()]
	if !ok {
		return false, false, false
	}

	return true, entry.isPinned, entry.isExplicit
}
