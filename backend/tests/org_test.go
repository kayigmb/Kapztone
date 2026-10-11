package tests

import (
	"strings"
	"testing"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/org"
)

func validRegistration() org.Registration {
	return org.Registration{
		Name: "Acme Coffee", Slug: fakeSlug, Kind: org.KindProducer,
		UserName: "Ada", Email: fakeEmail, Password: testPassword,
	}
}

func TestRegistrationValidate(t *testing.T) {
	ok := validRegistration()
	ok.Name, ok.Email = "  Acme Coffee  ", " "+fakeEmail+" "
	if msg := ok.Validate(); msg != "" {
		t.Fatalf("valid registration rejected: %s", msg)
	}
	if ok.Name != "Acme Coffee" || ok.Email != fakeEmail {
		t.Fatalf("input not trimmed: %+v", ok)
	}

	cases := map[string]func(*org.Registration){
		"empty name":        func(r *org.Registration) { r.Name = " " },
		"long name":         func(r *org.Registration) { r.Name = strings.Repeat("a", org.MaxNameLen+1) },
		"upper case slug":   func(r *org.Registration) { r.Slug = "Acme" },
		"reserved slug":     func(r *org.Registration) { r.Slug = auth.ReservedSlugs[0] },
		"slug with dot":     func(r *org.Registration) { r.Slug = "a.b" },
		"unknown kind":      func(r *org.Registration) { r.Kind = "bank" },
		"empty user name":   func(r *org.Registration) { r.UserName = "" },
		"bad email":         func(r *org.Registration) { r.Email = "not-an-email" },
		"email with name":   func(r *org.Registration) { r.Email = "Ada <" + fakeEmail + ">" },
		"short password":    func(r *org.Registration) { r.Password = strings.Repeat("a", auth.MinPasswordLen-1) },
		"too long password": func(r *org.Registration) { r.Password = strings.Repeat("a", auth.MaxPasswordLen+1) },
	}
	for name, breakIt := range cases {
		t.Run(name, func(t *testing.T) {
			r := validRegistration()
			breakIt(&r)
			if r.Validate() == "" {
				t.Fatal("accepted")
			}
		})
	}
}

func TestReviewValidate(t *testing.T) {
	cases := []struct {
		name   string
		review org.Review
		ok     bool
	}{
		{"verify", org.Review{Decision: org.StatusVerified}, true},
		{"reject with reason", org.Review{Decision: org.StatusRejected, Reason: "licence expired"}, true},
		{"reject without reason", org.Review{Decision: org.StatusRejected, Reason: "  "}, false},
		{"back to pending is not a decision", org.Review{Decision: org.StatusPending}, false},
		{"unknown decision", org.Review{Decision: "maybe"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.review.Validate() == ""; got != tc.ok {
				t.Fatalf("valid = %v, want %v", got, tc.ok)
			}
		})
	}
}
