package database

func TestGenerateAdvisoryLockId(t *testing.T) {
	id, err := p.generateAdvisoryLockId("database_name")
	if err != nil {
		t.Errorf("expected err to be nil, got %v", err)
	}
	if len(id) == 0 {
		t.Errorf("expected generated id not to be empty")
	}
	t.Logf("generated id: %v", id)
}
