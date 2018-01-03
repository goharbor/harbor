package pkcs11

// A test of using several pkcs11 sessions in parallel for signing across
// multiple goroutines. Access to the PKCS11 module is thread-safe because of
// the C.CKF_OS_LOCKING_OK param and nil mutex functions that the pkcs11
// package passes to C.Initialize, which indicate that the module should use OS
// locking primitives on its own.
//
// Note that while access to the module is thread-safe, sessions are not thread
// safe, and each session must be protected from simultaneous use by some
// synchronization mechanism. In this case we use a cache of sessions (as
// embodied by the `signer` struct), protected by a condition variable. So long
// as there is an available signer in the cache, it is popped off and used. If
// there are no signers available, the caller blocks until there is one
// available.
//
// Please set the appropiate env variables. See the init function.
import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
)

var (
	module          = "/usr/lib/softhsm/libsofthsm.so"
	tokenLabel      = "softhsm token"
	privateKeyLabel = "my key"
	pin             = "1234"
)

func init() {
	if x := os.Getenv("SOFTHSM_LIB"); x != "" {
		module = x
	}
	if x := os.Getenv("SOFTHSM_TOKENLABEL"); x != "" {
		tokenLabel = x
	}
	if x := os.Getenv("SOFTHSM_PRIVKEYLABEL"); x != "" {
		privateKeyLabel = x
	}
	if x := os.Getenv("SOFTHSM_PIN"); x != "" {
		pin = x
	}
	wd, _ := os.Getwd()
	os.Setenv("SOFTHSM_CONF", wd+"/softhsm.conf")
}

func initPKCS11Context(modulePath string) (*Ctx, error) {
	context := New(modulePath)

	if context == nil {
		return nil, fmt.Errorf("unable to load PKCS#11 module")
	}

	err := context.Initialize()
	return context, err
}

func getSlot(p *Ctx, label string) (uint, error) {
	slots, err := p.GetSlotList(true)
	if err != nil {
		return 0, err
	}
	for _, slot := range slots {
		_, err := p.GetSlotInfo(slot)
		if err != nil {
			return 0, err
		}
		tokenInfo, err := p.GetTokenInfo(slot)
		if err != nil {
			return 0, err
		}
		if tokenInfo.Label == label {
			return slot, nil
		}
	}
	return 0, fmt.Errorf("Slot not found: %s", label)
}

func getPrivateKey(context *Ctx, session SessionHandle, label string) (ObjectHandle, error) {
	var noKey ObjectHandle
	template := []*Attribute{
		NewAttribute(CKA_CLASS, CKO_PRIVATE_KEY),
		NewAttribute(CKA_LABEL, label),
	}
	if err := context.FindObjectsInit(session, template); err != nil {
		return noKey, err
	}
	objs, _, err := context.FindObjects(session, 2)
	if err != nil {
		return noKey, err
	}
	if err = context.FindObjectsFinal(session); err != nil {
		return noKey, err
	}

	if len(objs) == 0 {
		err = fmt.Errorf("private key not found")
		return noKey, err
	}
	return objs[0], nil
}

type signer struct {
	context    *Ctx
	session    SessionHandle
	privateKey ObjectHandle
}

func makeSigner(context *Ctx) (*signer, error) {
	slot, err := getSlot(context, tokenLabel)
	if err != nil {
		return nil, err
	}
	session, err := context.OpenSession(slot, CKF_SERIAL_SESSION)
	if err != nil {
		return nil, err
	}

	if err = context.Login(session, CKU_USER, pin); err != nil {
		context.CloseSession(session)
		return nil, err
	}

	privateKey, err := getPrivateKey(context, session, privateKeyLabel)
	if err != nil {
		context.CloseSession(session)
		return nil, err
	}
	return &signer{context, session, privateKey}, nil
}

func (s *signer) sign(input []byte) ([]byte, error) {
	mechanism := []*Mechanism{NewMechanism(CKM_RSA_PKCS, nil)}
	if err := s.context.SignInit(s.session, mechanism, s.privateKey); err != nil {
		log.Fatalf("SignInit: %s", err)
	}

	signed, err := s.context.Sign(s.session, input)
	if err != nil {
		log.Fatalf("Sign: %s", err)
	}
	return signed, nil
}

type cache struct {
	signers []*signer
	// this variable signals the condition that there are signers available to be
	// used.
	cond *sync.Cond
}

func newCache(signers []*signer) cache {
	var mutex sync.Mutex
	return cache{
		signers: signers,
		cond:    sync.NewCond(&mutex),
	}
}

func (c *cache) get() *signer {
	c.cond.L.Lock()
	for len(c.signers) == 0 {
		c.cond.Wait()
	}

	instance := c.signers[len(c.signers)-1]
	c.signers = c.signers[:len(c.signers)-1]
	c.cond.L.Unlock()
	return instance
}

func (c *cache) put(instance *signer) {
	c.cond.L.Lock()
	c.signers = append(c.signers, instance)
	c.cond.Signal()
	c.cond.L.Unlock()
}

func (c *cache) sign(input []byte) ([]byte, error) {
	instance := c.get()
	defer c.put(instance)
	return instance.sign(input)
}

// TODO(miek): disabled for now. Fill out the correct values in hsm.db so we can use it.
func testParallel(t *testing.T) {
	if module == "" || tokenLabel == "" || pin == "" || privateKeyLabel == "" {
		t.Fatal("Must pass all flags: module, tokenLabel, pin, and privateKeyLabel")
		return
	}

	context, err := initPKCS11Context(module)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		context.Finalize()
		context.Destroy()
	}()

	const nSigners = 100
	const nSignatures = 1000
	signers := make([]*signer, nSigners)
	for i := 0; i < nSigners; i++ {
		signers[i], err = makeSigner(context)
		if err != nil {
			t.Fatalf("Problem making signer: %s", err)
		}
	}
	pool := newCache(signers)

	output := make(chan []byte, nSignatures)
	for i := 0; i < nSignatures; i++ {
		go func() {
			result, err := pool.sign([]byte("hi"))
			if err != nil {
				t.Fatal(err)
			}
			output <- result
		}()
	}

	for i := 0; i < nSignatures; i++ {
		// Consume the output of the signers, but do nothing with it.
		<-output
	}

	for i := 0; i < nSigners; i++ {
		// Note: It is not necessary to call context.Logout. Closing the last
		// session will automatically log out, per PKCS#11 API.
		context.CloseSession(signers[i].session)
	}
}
