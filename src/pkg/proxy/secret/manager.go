package secret

import (
	"sync"
	"time"

	"github.com/goharbor/harbor/src/common/utils"
)

const defaultExpiration = 15 * time.Second

type targetRepository struct {
	name      string
	expiresAt time.Time
}

// Manager generates and verifies the secret for repositories under proxy project
// The secret normally is used for authorizing a request trying to push artifact to a project
// A secret can be used only once and expires in a short period of time.
// As the request will be sent to 127.0.0.1 so the secret will live in one process.
type Manager interface {
	// Generate generates a secret for the given repository, sample value for repository: "library/ubuntu"
	Generate(repository string) string
	// Verify verifies the secret against repo name, after the verification the secret should be invalid
	Verify(secret, repository string) bool
}

type mgr struct {
	m   *sync.Map
	exp time.Duration
}

func (man *mgr) Generate(rn string) string {
	s := utils.GenerateRandomStringWithLen(8)
	man.m.Store(s, targetRepository{name: rn, expiresAt: time.Now().Add(man.exp)})
	return s
}

func (man *mgr) Verify(sec, rn string) bool {
	v, ok := man.m.Load(sec)
	if !ok {
		return false
	}
	p, ok := v.(targetRepository)
	if ok && p.name == rn {
		defer man.m.Delete(sec)
		return p.expiresAt.After(time.Now())
	}
	return false
}

var (
	defaultManager Manager
	once           sync.Once
)

// GetManager returns the default manager which is a singleton in the package
func GetManager() Manager {
	once.Do(func() {
		defaultManager = createManager(defaultExpiration)
	})
	return defaultManager
}

func createManager(d time.Duration) Manager {
	return &mgr{
		m:   &sync.Map{},
		exp: d,
	}
}
