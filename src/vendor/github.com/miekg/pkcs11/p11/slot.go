package p11

import "github.com/miekg/pkcs11"

// Slot represents a slot that may hold a token.
type Slot struct {
	ctx *pkcs11.Ctx
	id  uint
}

// Info returns information about the Slot.
func (s Slot) Info() (pkcs11.SlotInfo, error) {
	return s.ctx.GetSlotInfo(s.id)
}

// TokenInfo returns information about the token in a Slot, if applicable.
func (s Slot) TokenInfo() (pkcs11.TokenInfo, error) {
	return s.ctx.GetTokenInfo(s.id)
}

// OpenSession opens a read-only session with the token in this slot.
func (s Slot) OpenSession() (Session, error) {
	return s.openSession(0)
}

// OpenWriteSession opens a read-write session with the token in this slot.
func (s Slot) OpenWriteSession() (Session, error) {
	return s.openSession(pkcs11.CKF_RW_SESSION)
}

func (s Slot) openSession(flags uint) (Session, error) {
	// CKF_SERIAL_SESSION is always mandatory for legacy reasons, per PKCS#11.
	handle, err := s.ctx.OpenSession(s.id, flags|pkcs11.CKF_SERIAL_SESSION)
	if err != nil {
		return nil, err
	}
	return &sessionImpl{
		ctx:    s.ctx,
		handle: handle,
	}, nil
}

// CloseAllSessions closes all sessions on this slot.
func (s Slot) CloseAllSessions() error {
	return s.ctx.CloseAllSessions(s.id)
}

// Mechanisms returns a list of Mechanisms available on the token in this
// slot.
func (s Slot) Mechanisms() ([]Mechanism, error) {
	list, err := s.ctx.GetMechanismList(s.id)
	if err != nil {
		return nil, err
	}
	result := make([]Mechanism, len(list))
	for i, mech := range list {
		result[i] = Mechanism{
			mechanism: mech,
			slot:      s,
		}
	}
	return result, nil
}

// InitToken initializes the token in this slot, setting its label to
// tokenLabel. If the token was not previously initialized, its security officer
// PIN is set to the provided string. If the token is already initialized, the
// provided PIN will be checked against the existing security officer PIN, and
// the token will only be reinitialized if there is a match.
//
// According to PKCS#11: "When a token is initialized, all objects that can be
// destroyed are destroyed (i.e., all except for 'indestructible' objects such
// as keys built into the token). Also, access by the normal user is disabled
// until the SO sets the normal user’s PIN."
func (s Slot) InitToken(securityOfficerPIN string, tokenLabel string) error {
	return s.ctx.InitToken(s.id, securityOfficerPIN, tokenLabel)
}

// ID returns the slot's ID.
func (s Slot) ID() uint {
	return s.id
}

// Mechanism represents a cipher, signature algorithm, hash function, or other
// function that a token can perform.
type Mechanism struct {
	mechanism *pkcs11.Mechanism
	slot      Slot
}

// Info returns information about this mechanism.
func (m *Mechanism) Info() (pkcs11.MechanismInfo, error) {
	return m.slot.ctx.GetMechanismInfo(m.slot.id, []*pkcs11.Mechanism{m.mechanism})
}
