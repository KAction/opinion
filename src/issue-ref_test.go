package src

import "testing"

func TestIssueRefSet(t *testing.T) {
	value := IssueRef{}
	s := "foo/bar#10"

	err := value.Set(s)
	if err != nil {
		t.Errorf("Failed to parse `%s'", s)
	}
	if value.Owner != "foo" {
		t.Errorf("Bad value of Owner: %s != `foo'", value.Owner)
	}
	if value.Repo != "bar" {
		t.Errorf("Bad value of Repo: `%s' != `bar'", value.Repo)
	}
	if value.Number != 10 {
		t.Errorf("Bad value of Number: `%d' != 10", value.Number)
	}

	s = "foo/bar#bzr"
	err = value.Set(s)
	if err == nil {
		t.Errorf("Failed to report error about `%s'", s)
	}
}

func TestIssueRefShow(t *testing.T) {
	value := IssueRef{Owner: "kaction", Repo: "opinion", Number: 42}
	actual := value.String()
	expected := "kaction/opinion#42"

	if actual != expected {
		t.Errorf("Failed to format value: `%s' != `%x' (expected)",
			actual, expected)
	}
}
