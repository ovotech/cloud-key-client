package keys

import (
	"testing"
)

func TestValidateGcpProjectString(t *testing.T) {
	if validateGcpProjectString("") == nil {
		t.Error("The code did not error")
	}
}

func TestGcpProjectName(t *testing.T) {
	actual := gcpProjectName("project")
	expected := "projects/project"
	if expected != actual {
		t.Errorf("Incorrect string returned, got: %s, want: %s.", actual, expected)
	}
}

func TestGcpServiceAccountName(t *testing.T) {
	actual := gcpServiceAccountName("project", "sa")
	expected := "projects/project/serviceAccounts/sa"
	if expected != actual {
		t.Errorf("Incorrect string returned, got: %s, want: %s.", actual, expected)
	}
}

func TestGcpServiceAccountKeyName(t *testing.T) {
	actual := gcpServiceAccountKeyName("project", "sa", "key")
	expected := "projects/project/serviceAccounts/sa/keys/key"
	if expected != actual {
		t.Errorf("Incorrect string returned, got: %s, want: %s.", actual, expected)
	}
}
