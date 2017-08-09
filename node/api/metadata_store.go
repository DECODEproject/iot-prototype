package api

import "sync"

type Metadata struct {
	Description string `json:"description" description:"human readable description of the data"`
	Subject     string `json:"subject" description:"unique name for the data"`
}

func NewMetadataStore() *MetadataStore {
	return &MetadataStore{
		store: map[string]Metadata{},
	}
}

type MetadataStore struct {
	lock  sync.RWMutex
	store map[string]Metadata
}

func (m *MetadataStore) Add(meta Metadata) {

	m.lock.Lock()
	defer m.lock.Unlock()
	m.store[meta.Subject] = meta
}

func (m *MetadataStore) FindBySubject(subject string) Metadata {

	m.lock.Lock()
	defer m.lock.Unlock()
	return m.store[subject]
}

func (m *MetadataStore) All() []Metadata {

	m.lock.Lock()
	defer m.lock.Unlock()

	list := []Metadata{}
	for _, each := range m.store {
		list = append(list, each)
	}

	return list

}