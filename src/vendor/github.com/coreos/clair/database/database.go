// Copyright 2015 clair authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package database defines the Clair's models and a common interface for database implementations.
package database

import (
	"errors"
	"time"
)

var (
	// ErrBackendException is an error that occurs when the database backend does
	// not work properly (ie. unreachable).
	ErrBackendException = errors.New("database: an error occured when querying the backend")

	// ErrInconsistent is an error that occurs when a database consistency check
	// fails (ie. when an entity which is supposed to be unique is detected twice)
	ErrInconsistent = errors.New("database: inconsistent database")

	// ErrCantOpen is an error that occurs when the database could not be opened
	ErrCantOpen = errors.New("database: could not open database")
)

// Datastore is the interface that describes a database backend implementation.
type Datastore interface {
	// # Namespace
	// ListNamespaces returns the entire list of known Namespaces.
	ListNamespaces() ([]Namespace, error)

	// # Layer
	// InsertLayer stores a Layer in the database.
	// A Layer is uniquely identified by its Name. The Name and EngineVersion fields are mandatory.
	// If a Parent is specified, it is expected that it has been retrieved using FindLayer.
	// If a Layer that already exists is inserted and the EngineVersion of the given Layer is higher
	// than the stored one, the stored Layer should be updated.
	// The function has to be idempotent, inserting a layer that already exists shouln'd return an
	// error.
	InsertLayer(Layer) error

	// FindLayer retrieves a Layer from the database.
	// withFeatures specifies whether the Features field should be filled. When withVulnerabilities is
	// true, the Features field should be filled and their AffectedBy fields should contain every
	// vulnerabilities that affect them.
	FindLayer(name string, withFeatures, withVulnerabilities bool) (Layer, error)

	// DeleteLayer deletes a Layer from the database and every layers that are based on it,
	// recursively.
	DeleteLayer(name string) error

	// # Vulnerability
	// ListVulnerabilities returns the list of vulnerabilies of a certain Namespace.
	// The Limit and page parameters are used to paginate the return list.
	// The first given page should be 0. The function will then return the next available page.
	// If there is no more page, -1 has to be returned.
	ListVulnerabilities(namespaceName string, limit int, page int) ([]Vulnerability, int, error)

	// InsertVulnerabilities stores the given Vulnerabilities in the database, updating them if
	// necessary. A vulnerability is uniquely identified by its Namespace and its Name.
	// The FixedIn field may only contain a partial list of Features that are affected by the
	// Vulnerability, along with the version in which the vulnerability is fixed. It is the
	// responsibility of the implementation to update the list properly. A version equals to
	// types.MinVersion means that the given Feature is not being affected by the Vulnerability at
	// all and thus, should be removed from the list. It is important that Features should be unique
	// in the FixedIn list. For example, it doesn't make sense to have two `openssl` Feature listed as
	// a Vulnerability can only be fixed in one Version. This is true because Vulnerabilities and
	// Features are Namespaced (i.e. specific to one operating system).
	// Each vulnerability insertion or update has to create a Notification that will contain the
	// old and the updated Vulnerability, unless createNotification equals to true.
	InsertVulnerabilities(vulnerabilities []Vulnerability, createNotification bool) error

	// FindVulnerability retrieves a Vulnerability from the database, including the FixedIn list.
	FindVulnerability(namespaceName, name string) (Vulnerability, error)

	// DeleteVulnerability removes a Vulnerability from the database.
	// It has to create a Notification that will contain the old Vulnerability.
	DeleteVulnerability(namespaceName, name string) error

	// InsertVulnerabilityFixes adds new FixedIn Feature or update the Versions of existing ones to
	// the specified Vulnerability in the database.
	// It has has to create a Notification that will contain the old and the updated Vulnerability.
	InsertVulnerabilityFixes(vulnerabilityNamespace, vulnerabilityName string, fixes []FeatureVersion) error

	// DeleteVulnerabilityFix removes a FixedIn Feature from the specified Vulnerability in the
	// database. It can be used to store the fact that a Vulnerability no longer affects the given
	// Feature in any Version.
	// It has has to create a Notification that will contain the old and the updated Vulnerability.
	DeleteVulnerabilityFix(vulnerabilityNamespace, vulnerabilityName, featureName string) error

	// # Notification
	// GetAvailableNotification returns the Name, Created, Notified and Deleted fields of a
	// Notification that should be handled. The renotify interval defines how much time after being
	// marked as Notified by SetNotificationNotified, a Notification that hasn't been deleted should
	// be returned again by this function. A Notification for which there is a valid Lock with the
	// same Name should not be returned.
	GetAvailableNotification(renotifyInterval time.Duration) (VulnerabilityNotification, error)

	// GetNotification returns a Notification, including its OldVulnerability and NewVulnerability
	// fields. On these Vulnerabilities, LayersIntroducingVulnerability should be filled with
	// every Layer that introduces the Vulnerability (i.e. adds at least one affected FeatureVersion).
	// The Limit and page parameters are used to paginate LayersIntroducingVulnerability. The first
	// given page should be VulnerabilityNotificationFirstPage. The function will then return the next
	// availage page. If there is no more page, NoVulnerabilityNotificationPage has to be returned.
	GetNotification(name string, limit int, page VulnerabilityNotificationPageNumber) (VulnerabilityNotification, VulnerabilityNotificationPageNumber, error)

	// SetNotificationNotified marks a Notification as notified and thus, makes it unavailable for
	// GetAvailableNotification, until the renotify duration is elapsed.
	SetNotificationNotified(name string) error

	// DeleteNotification marks a Notification as deleted, and thus, makes it unavailable for
	// GetAvailableNotification.
	DeleteNotification(name string) error

	// # Key/Value
	// InsertKeyValue stores or updates a simple key/value pair in the database.
	InsertKeyValue(key, value string) error

	// GetKeyValue retrieves a value from the database from the given key.
	// It returns an empty string if there is no such key.
	GetKeyValue(key string) (string, error)

	// # Lock
	// Lock creates or renew a Lock in the database with the given name, owner and duration.
	// After the specified duration, the Lock expires by itself if it hasn't been unlocked, and thus,
	// let other users create a Lock with the same name. However, the owner can renew its Lock by
	// setting renew to true. Lock should not block, it should instead returns whether the Lock has
	// been successfully acquired/renewed. If it's the case, the expiration time of that Lock is
	// returned as well.
	Lock(name string, owner string, duration time.Duration, renew bool) (bool, time.Time)

	// Unlock releases an existing Lock.
	Unlock(name, owner string)

	// FindLock returns the owner of a Lock specified by the name, and its experation time if it
	// exists.
	FindLock(name string) (string, time.Time, error)

	// # Miscellaneous
	// Ping returns the health status of the database.
	Ping() bool

	// Close closes the database and free any allocated resource.
	Close()
}
