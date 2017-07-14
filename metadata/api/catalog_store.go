package api

import (
	"errors"
	"sync"
)

var (
	ErrLocationNotExists = errors.New("location does not exist")
)

type MetadataStore struct {
	Items     itemMap
	Locations locationMap
}

func NewMetadataStore() *MetadataStore {
	return &MetadataStore{
		Items:     newItemMap(),
		Locations: newLocationMap(),
	}
}
func (m *MetadataStore) All() []ItemWithLocation {

	list := []ItemWithLocation{}
	for _, each := range m.Items.All() {
		location := m.Locations.Get(each.LocationUID)
		list = append(list, ItemWithLocation{
			each,
			location,
		})
	}

	return list
}

type itemMap struct {
	lock  sync.RWMutex
	store map[string]CatalogItem
}

func newItemMap() itemMap {
	return itemMap{
		store: map[string]CatalogItem{},
	}
}

func (e itemMap) Add(i CatalogItem) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.store[i.UID] = i

}

func (e itemMap) Delete(uid string) {
	e.lock.Lock()
	defer e.lock.Unlock()

	delete(e.store, uid)
}

func (e itemMap) All() []CatalogItem {
	e.lock.Lock()
	defer e.lock.Unlock()

	list := []CatalogItem{}
	for _, each := range e.store {
		list = append(list, each)
	}

	return list

}

type locationMap struct {
	lock  sync.RWMutex
	store map[string]Location
}

func newLocationMap() locationMap {
	return locationMap{
		store: map[string]Location{},
	}
}

func (e locationMap) Add(i Location) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.store[i.UID] = i
}

func (e locationMap) Get(uid string) Location {
	e.lock.Lock()
	defer e.lock.Unlock()

	return e.store[uid]
}

func (e locationMap) Exists(uid string) bool {
	e.lock.Lock()
	defer e.lock.Unlock()

	_, found := e.store[uid]
	return found
}

func (e locationMap) Replace(uid string, i Location) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	_, found := e.store[uid]

	if !found {
		return ErrLocationNotExists
	}

	e.store[uid] = i
	return nil
}
