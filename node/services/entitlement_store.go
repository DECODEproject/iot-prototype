package services

import (
	"errors"
	"sync"
)

type EntitlementStore struct {
	Accepted  entitlementMap
	Declined  entitlementMap
	Requested entitlementMap
	Revoked   entitlementMap
}

func NewEntitlementStore() *EntitlementStore {
	return &EntitlementStore{
		Accepted:  entitlementMap{},
		Declined:  entitlementMap{},
		Requested: entitlementMap{},
		Revoked:   entitlementMap{},
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

	return f(&ent)
}

func (e entitlementMap) Get(uid string) (Entitlement, bool) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	ent, found := e.store[uid]
	return ent, found
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

	return list
}
