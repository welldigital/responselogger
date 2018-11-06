package urlpattern

import "testing"

func Test(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{
			url:      "/pharmacy/user/ded4c637-8fed-4ac2-9215-4b41294febef/requestsandorders",
			expected: "/pharmacy/user/{uuid}/requestsandorders",
		},
		{
			url:      "/pharmacy/request/3191/reject",
			expected: "/pharmacy/request/{integer}/reject",
		},
	}
	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			actual := Extract(tc.url)
			if actual != tc.expected {
				t.Errorf("expected '%s', got '%s'", tc.expected, actual)
			}
		})
	}
}
