package api

import (
	"errors"
	"sort"
	"sync"

	"gogs.dyne.org/DECODE/decode-prototype-da/utils"
)

type EntitlementStore struct {
	Accepted  entitlementMap
	Declined  entitlementMap
	Requested entitlementMap
	Revoked   entitlementMap
}

func NewEntitlementStore() *EntitlementStore {
	return &EntitlementStore{
		Accepted:  newEntitlementMap(),
		Declined:  newEntitlementMap(),
		Requested: newEntitlementMap(),
		Revoked:   newEntitlementMap(),
	}
}

func newEntitlementMap() entitlementMap {
	return entitlementMap{
		store: map[string]Entitlement{},
	}
}

type entitlementMap struct {
	lock  sync.RWMutex
	store map[string]Entitlement
}

type entitlementUpdater func(e *Entitlement) error

var (
	ErrEntitlementNotFound = errors.New("entitlement not found")
)

func (e entitlementMap) Update(uid string, f entitlementUpdater) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	ent, found := e.store[uid]

	if !found {
		return ErrEntitlementNotFound
	}

	err := f(&ent)

	if err != nil {
		return err
	}

	e.store[uid] = ent

	return nil
}

func (e entitlementMap) Get(uid string) (Entitlement, bool) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	ent, found := e.store[uid]
	return ent, found
}

// FindForSubject looks for an entitlement for the data using it's subject key.
// A subject key looks like this - data://foo/bar/xxx
// and has the form data:// and then a folder type path of an unknown number of segments.
// FindForSubject will look for an 'entitlement' for the exact key then traverse the path
// up the chain until it finds an 'entitlement'.
func (e entitlementMap) FindForSubject(subject utils.Subject) (Entitlement, bool) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	perms := subject.Perms()

	// TODO : optimise optimise optimise - rewrite as a trie
	for _, s := range perms {
		for _, ent := range e.store {
			if ent.Subject == s {
				ffs := ent
				return ffs, true
			}
		}
	}

	return Entitlement{}, false
}

func (e entitlementMap) Add(ent Entitlement) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.store[ent.UID] = ent
}

func (e entitlementMap) Delete(uid string) {
	e.lock.Lock()
	defer e.lock.Unlock()

	delete(e.store, uid)
}

func (e entitlementMap) All() []Entitlement {
	e.lock.RLock()
	defer e.lock.RUnlock()

	list := []Entitlement{}
	for _, each := range e.store {
		list = append(list, each)
	}

	sort.Sort(byEntSubject(list))

	return list
}

func (e entitlementMap) AppendOrReplaceOnSubject(ent Entitlement) {
	e.lock.Lock()
	defer e.lock.Unlock()

	found := false
	existing := Entitlement{}

	for _, e := range e.store {
		if e.Subject == ent.Subject {
			found = true
			existing = e
		}

	}

	if found {
		ent.UID = existing.UID
		e.store[existing.UID] = ent
		return
	}

	e.store[ent.UID] = ent

}

type byEntSubject []Entitlement

func (a byEntSubject) Len() int           { return len(a) }
func (a byEntSubject) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byEntSubject) Less(i, j int) bool { return a[i].Subject < a[j].Subject }
