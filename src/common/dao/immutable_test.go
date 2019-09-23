package dao

import (
	"testing"

	"github.com/goharbor/harbor/src/common/models"
)

func TestCreateImmutableRule(t *testing.T) {
	ir := &models.ImmutableRule{TagFilter: "**", ProjectID: 1}
	id, err := CreateImmutableRule(ir)
	if err != nil {
		t.Errorf("error: %+v", err)
	}
	if id <= 0 {
		t.Error("Can not create immutable tag rule")
	}
	_, err = DeleteImmutableRule(id)
	if err != nil {
		t.Errorf("error: %+v", err)
	}
}

func TestUpdateImmutableRule(t *testing.T) {
	ir := &models.ImmutableRule{TagFilter: "**", ProjectID: 1}
	id, err := CreateImmutableRule(ir)
	if err != nil {
		t.Errorf("error: %+v", err)
	}
	if id <= 0 {
		t.Error("Can not create immutable tag rule")
	}

	updatedIR := &models.ImmutableRule{ID: id, TagFilter: "1.2.0", ProjectID: 1}
	updatedCnt, err := UpdateImmutableRule(1, updatedIR)
	if err != nil {
		t.Errorf("error: %+v", err)
	}
	if updatedCnt <= 0 {
		t.Error("Failed to update immutable id")
	}

	newIr, err := GetImmutableRule(id)

	if err != nil {
		t.Errorf("error: %+v", err)
	}

	if newIr.TagFilter != "1.2.0" {
		t.Error("Failed to update immutable tag")
	}

	defer DeleteImmutableRule(id)

}

func TestEnableImmutableRule(t *testing.T) {
	ir := &models.ImmutableRule{TagFilter: "**", ProjectID: 1}
	id, err := CreateImmutableRule(ir)
	if err != nil {
		t.Errorf("error: %+v", err)
	}
	if id <= 0 {
		t.Error("Can not create immutable tag rule")
	}

	ToggleImmutableRule(id, false)
	newIr, err := GetImmutableRule(id)

	if err != nil {
		t.Errorf("error: %+v", err)
	}

	if newIr.Enabled != false {
		t.Error("Failed to disable the immutable rule")
	}

	defer DeleteImmutableRule(id)
}

func TestGetImmutableRuleByProject(t *testing.T) {
	irs := []*models.ImmutableRule{
		{TagFilter: "version1", ProjectID: 99},
		{TagFilter: "version2", ProjectID: 99},
		{TagFilter: "version3", ProjectID: 99},
		{TagFilter: "version4", ProjectID: 99},
	}
	for _, ir := range irs {
		CreateImmutableRule(ir)
	}

	qrs, err := QueryImmutableRuleByProjectID(99)
	if err != nil {
		t.Errorf("error: %+v", err)
	}

	if len(qrs) != 4 {
		t.Error("Failed to query 4 rows!")
	}

	defer ExecuteBatchSQL([]string{"delete from immutable_tag_rule where project_id = 99 "})

}
func TestGetEnabledImmutableRuleByProject(t *testing.T) {
	irs := []*models.ImmutableRule{
		{TagFilter: "version1", ProjectID: 99},
		{TagFilter: "version2", ProjectID: 99},
		{TagFilter: "version3", ProjectID: 99},
		{TagFilter: "version4", ProjectID: 99},
	}
	for i, ir := range irs {
		id, _ := CreateImmutableRule(ir)
		if i == 1 {
			ToggleImmutableRule(id, false)
		}

	}

	qrs, err := QueryEnabledImmutableRuleByProjectID(99)
	if err != nil {
		t.Errorf("error: %+v", err)
	}

	if len(qrs) != 3 {
		t.Errorf("Failed to query 3 rows!, got %v", len(qrs))
	}

	defer ExecuteBatchSQL([]string{"delete from immutable_tag_rule where project_id = 99 "})

}
