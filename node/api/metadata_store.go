package api

import (
	"sort"
	"sync"

	"gogs.dyne.org/DECODE/decode-prototype-da/utils"
)

type Metadata struct {
	Description string `json:"description" description:"human readable description of the data"`
	Subject     string `json:"subject" description:"description of the data"`
	Name        string `json:"name" description:"name of the data"`
	Path        string `json:"path" description:"path to the key of the data"`
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

func (m *MetadataStore) FindBySubject(subject utils.Subject) (Metadata, bool) {

	m.lock.Lock()
	defer m.lock.Unlock()

	perms := subject.Perms()

	// TODO : optimise optimise optimise - rewrite as a trie
	for _, s := range perms {
		for _, m := range m.store {
			if m.Subject == s {
				ffs := m
				return ffs, true
			}
		}
	}
	return Metadata{}, false
}

func (m *MetadataStore) All() []Metadata {

	m.lock.Lock()
	defer m.lock.Unlock()

	list := []Metadata{}
	for _, each := range m.store {

		sub, _ := utils.ParseSubject(each.Subject)

		if !sub.IsRoot() {
			list = append(list, each)
		}
	}

	sort.Sort(byMetSubject(list))
	return list

}

type byMetSubject []Metadata

func (a byMetSubject) Len() int           { return len(a) }
func (a byMetSubject) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byMetSubject) Less(i, j int) bool { return a[i].Subject < a[j].Subject }
