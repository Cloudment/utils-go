package utils

import (
	"testing"
)

func TestValidatePagination(t *testing.T) {
	t.Parallel()
	tests := []struct {
		page, limit, expectedPage, expectedLimit int
	}{
		{1, 50, 1, 50},
		{0, 0, 0, 10},
		{5, 200, 5, 100},
		{-1, 20, 0, 20},
		{2, -10, 2, -10},
	}

	for _, test := range tests {
		page, limit := ValidatePagination(test.page, test.limit)
		if page != test.expectedPage || limit != test.expectedLimit {
			t.Errorf("ValidatePagination(%d, %d) = (%d, %d); want (%d, %d)", test.page, test.limit, page, limit, test.expectedPage, test.expectedLimit)
		}
	}
}

func BenchmarkValidatePagination(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidatePagination(1, 50)
	}
}

func TestToAnySlice(t *testing.T) {
	t.Parallel()

	in1 := []int{0, 1, 2, 3}
	var in2 []int

	out1 := ToAnySlice(in1)
	out2 := ToAnySlice(in2)

	expectedOut1 := []any{0, 1, 2, 3}
	expectedOut2 := []any{} // If this is changed, the test will fail

	if !IsEqual(out1, expectedOut1) {
		t.Errorf("Expected %v, got %v", expectedOut1, out1)
	}
	if !IsEqual(out2, expectedOut2) {
		t.Errorf("Expected %v, got %v", expectedOut2, out2)
	}
}

// TODO: Add better user agent testing, more comprehensive list of user agents and operating systems
type UserAgents struct {
	Desktop []string `json:"desktop"`
	Mobile  []string `json:"mobile"`
	Unknown []string `json:"unknown"`
}

var userAgents = UserAgents{
	Desktop: []string{
		// Appears that the Mac platform always states Intel regardless of the actual CPU
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.3",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 14.2; rv:115.0) Gecko/20100101 Firefox/115.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OPR/106.0.0.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_2_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OPR/106.0.0.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.3",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; Xbox; Xbox One) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edge/44.18363.8131",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/117.",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:115.0) Gecko/20100101 Firefox/115.0",
		"Mozilla/5.0 (X11; Linux i686; rv:109.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (X11; Linux i686; rv:115.0) Gecko/20100101 Firefox/115.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (X11; Ubuntu; Linux i686; rv:109.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:115.0) Gecko/20100101 Firefox/115.0",
	},
	Mobile: []string{
		"Mozilla/5.0 (Android 14; Mobile; LG-M255; rv:121.0) Gecko/121.0 Firefox/121.0",
		"Mozilla/5.0 (Android 14; Mobile; rv:109.0) Gecko/121.0 Firefox/121.0",
		"Mozilla/5.0 (Linux; Android 10; HD1913) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.210 Mobile Safari/537.36 EdgA/120.0.2210.126",
		"Mozilla/5.0 (Linux; Android 10; JNY-LX1; HMSCore 6.13.0.302) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.88 HuaweiBrowser/14.0.2.311 Mobile Safari/537.3",
		"Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Mobile Safari/537.3",
		"Mozilla/5.0 (Linux; Android 10; MED-LX9N; HMSCore 6.13.0.301) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.88 HuaweiBrowser/14.0.2.311 Mobile Safari/537.3",
		"Mozilla/5.0 (Linux; Android 10; ONEPLUS A6003) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.210 Mobile Safari/537.36 EdgA/120.0.2210.126",
		"Mozilla/5.0 (Linux; Android 10; Pixel 3 XL) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.210 Mobile Safari/537.36 EdgA/120.0.2210.126",
		"Mozilla/5.0 (Linux; Android 10; SM-G970F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.210 Mobile Safari/537.36 OPR/76.2.4027.73374",
		"Mozilla/5.0 (Linux; Android 10; YAL-L21) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.58 Mobile Safari/537.3",
		"Mozilla/5.0 (Linux; Android 11; SAMSUNG SM-A715F) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/23.0 Chrome/115.0.0.0 Mobile Safari/537.3",
		"Mozilla/5.0 (Linux; Android 14; SM-A546B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.6045.193 Mobile Safari/537.36 OPR/79.5.4195.7698",
		"Mozilla/5.0 (Linux; Android 14; SM-S901B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.6045.193 Mobile Safari/537.36 OPR/79.5.4195.7698",
		"Mozilla/5.0 (Linux; Android 9; JAT-L41) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Mobile Safari/537.3",
		"Mozilla/5.0 (iPad; CPU OS 14_2_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) FxiOS/121.0 Mobile/15E148 Safari/605.1.15",
		"Mozilla/5.0 (iPad; CPU OS 17_2_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 14_2_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) FxiOS/121.0 Mobile/15E148 Safari/605.1.15",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 15_8 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.6.6 Mobile/15E148 Safari/604.",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/120.0.6099.119 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) GSA/298.0.595435837 Mobile/15E148 Safari/604.",
	},
	Unknown: []string{
		"Broken",
	},
}

func TestGetOperatingSystemFromUserAgent(t *testing.T) {
	t.Parallel()
	for _, userAgent := range userAgents.Desktop {
		os := GetOperatingSystemFromUserAgent(userAgent)
		if os == "Unknown" {
			t.Errorf("Expected %s to be a known operating system", userAgent)
		}
	}
	for _, userAgent := range userAgents.Mobile {
		os := GetOperatingSystemFromUserAgent(userAgent)
		if os == "Unknown" {
			t.Errorf("Expected %s to be a known operating system", userAgent)
		}
	}
	for _, userAgent := range userAgents.Unknown {
		os := GetOperatingSystemFromUserAgent(userAgent)
		if os != "Unknown" {
			t.Errorf("Expected %s to be an unknown operating system", userAgent)
		}
	}
}

func BenchmarkGetOperatingSystemFromUserAgent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetOperatingSystemFromUserAgent(userAgents.Desktop[0])
	}
}

func TestIsEqual(t *testing.T) {
	t.Parallel()

	if !IsEqual(1, 1) {
		t.Error("Expected IsEqual(1, 1) to be true")
	}
	if IsEqual(1, 2) {
		t.Error("Expected IsEqual(1, 2) to be false")
	}

	if !IsEqual("hello", "hello") {
		t.Error("Expected IsEqual(\"hello\", \"hello\") to be true")
	}
	if IsEqual("hello", "world") {
		t.Error("Expected IsEqual(\"hello\", \"world\") to be false")
	}

	if !IsEqual(1.0, 1.0) {
		t.Error("Expected IsEqual(1.0, 1.0) to be true")
	}
	if IsEqual(1.0, 2.0) {
		t.Error("Expected IsEqual(1.0, 2.0) to be false")
	}

	if !IsEqual(true, true) {
		t.Error("Expected IsEqual(true, true) to be true")
	}
	if IsEqual(true, false) {
		t.Error("Expected IsEqual(true, false) to be false")
	}

	type testStruct struct {
		Name string
		Age  int
	}

	a := testStruct{Name: "John", Age: 30}
	b := testStruct{Name: "John", Age: 30}
	if !IsEqual(a, b) {
		t.Error("Expected IsEqual(testStruct{Name: \"John\", Age: 30}, testStruct{Name: \"John\", Age: 30}) to be true")
	}
	b.Age = 31
	if IsEqual(a, b) {
		t.Error("Expected IsEqual(testStruct{Name: \"John\", Age: 30}, testStruct{Name: \"John\", Age: 31}) to be false")
	}

	if !IsEqual(nil, nil) {
		t.Error("Expected IsEqual(nil, nil) to be true")
	}
	if IsEqual(nil, 1) {
		t.Error("Expected IsEqual(nil, 1) to be false")
	}
}
