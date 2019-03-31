package slack

import "testing"

func Test_validateEnums(t *testing.T) {

	cases := []struct {
		Value          string
		ExpectErrCount int
	}{
		{
			Value:          "foo",
			ExpectErrCount: 0,
		},
		{
			Value:          "bar",
			ExpectErrCount: 0,
		},
		{
			Value:          "baz",
			ExpectErrCount: 0,
		},
		{
			Value:          "none",
			ExpectErrCount: 1,
		},
	}

	validationFunc := validateEnums([]string{"foo", "bar", "baz"})

	for _, tc := range cases {
		_, errors := validationFunc(tc.Value, "test_arg")

		if len(errors) != tc.ExpectErrCount {
			t.Fatalf("Expected 1 validation error but %d", tc.ExpectErrCount)
		}
	}
}

func Test_containsAny(t *testing.T) {
	slices := []string{"foo", "bar", "baz"}

	cases := []struct {
		Value  string
		Expect bool
	}{
		{
			Value:  "foo",
			Expect: true,
		},
		{
			Value:  "bar",
			Expect: true,
		},
		{
			Value:  "baz",
			Expect: true,
		},
		{
			Value:  "none",
			Expect: false,
		},
	}

	for _, tc := range cases {
		actual := containsAny(slices, tc.Value)

		if actual != tc.Expect {
			t.Fatalf("Expected %t but %t", tc.Expect, actual)
		}
	}
}
