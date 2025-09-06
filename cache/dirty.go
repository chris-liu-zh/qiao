package cache

func (c *Cache) setDirty() {
	c.DirtyTotal++
}

func (c *Cache) clearDirty() {
	c.DirtyTotal = 0
}
