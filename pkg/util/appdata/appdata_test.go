package appdata_test

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"testing"
	"unicode"

	"git.indra-labs.org/dev/ind/pkg/util/appdata"
)

// TestAppDataDir tests the API for Dir to ensure it gives expected results for various operating systems.
func TestAppDataDir(t *testing.T) {
	// App name plus upper and lowercase variants.
	appName := "myapp"
	appNameUpper := string(unicode.ToUpper(rune(appName[0]))) + appName[1:]
	appNameLower := string(unicode.ToLower(rune(appName[0]))) + appName[1:]
	// When we're on Windows, set the expected local and roaming directories per the environment vars.  When we aren't
	// on Windows, the function should return the current directory when forced to provide the Windows path since the
	// environment variables won't exist.
	winLocal := "."
	winRoaming := "."
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		roamingAppData := os.Getenv("APPDATA")
		if localAppData == "" {
			localAppData = roamingAppData
		}
		winLocal = filepath.Join(localAppData, appNameUpper)
		winRoaming = filepath.Join(roamingAppData, appNameUpper)
	}
	// Get the home directory to use for testing expected results.
	var homeDir string
	usr, e := user.Current()
	if e != nil {
		t.Errorf("user.Current: %v", e)
		return
	}
	homeDir = usr.HomeDir
	// Mac node data directory.
	macAppData := filepath.Join(homeDir, "Library", "Application Support")
	tests := []struct {
		goos    string
		appName string
		roaming bool
		want    string
	}{
		// Various combinations of application name casing, leading period, operating system, and roaming flags.
		{"windows", appNameLower, false, winLocal},
		{"windows", appNameUpper, false, winLocal},
		{"windows", "." + appNameLower, false, winLocal},
		{"windows", "." + appNameUpper, false, winLocal},
		{"windows", appNameLower, true, winRoaming},
		{"windows", appNameUpper, true, winRoaming},
		{"windows", "." + appNameLower, true, winRoaming},
		{"windows", "." + appNameUpper, true, winRoaming},
		{
			"linux",
			appNameLower,
			false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"linux",
			appNameUpper,
			false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"linux", "." + appNameLower, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"linux", "." + appNameUpper, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"darwin",
			appNameLower,
			false,
			filepath.Join(macAppData, appNameUpper),
		},
		{
			"darwin",
			appNameUpper,
			false,
			filepath.Join(macAppData, appNameUpper),
		},
		{
			"darwin", "." + appNameLower, false,
			filepath.Join(macAppData, appNameUpper),
		},
		{
			"darwin", "." + appNameUpper, false,
			filepath.Join(macAppData, appNameUpper),
		},
		{
			"openbsd", appNameLower, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"openbsd", appNameUpper, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"openbsd", "." + appNameLower, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"openbsd", "." + appNameUpper, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"freebsd", appNameLower, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"freebsd", appNameUpper, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"freebsd", "." + appNameLower, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"freebsd", "." + appNameUpper, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"netbsd", appNameLower, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"netbsd", appNameUpper, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"netbsd", "." + appNameLower, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"netbsd", "." + appNameUpper, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{"plan9", appNameLower, false, filepath.Join(homeDir, appNameLower)},
		{"plan9", appNameUpper, false, filepath.Join(homeDir, appNameLower)},
		{
			"plan9", "." + appNameLower, false,
			filepath.Join(homeDir, appNameLower),
		},
		{
			"plan9", "." + appNameUpper, false,
			filepath.Join(homeDir, appNameLower),
		},
		{
			"unrecognized", appNameLower, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"unrecognized", appNameUpper, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"unrecognized", "." + appNameLower, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		{
			"unrecognized", "." + appNameUpper, false,
			filepath.Join(homeDir, "."+appNameLower),
		},
		// Node application name provided, so expect current directory.
		{"windows", "", false, "."},
		{"windows", "", true, "."},
		{"linux", "", false, "."},
		{"darwin", "", false, "."},
		{"openbsd", "", false, "."},
		{"freebsd", "", false, "."},
		{"netbsd", "", false, "."},
		{"plan9", "", false, "."},
		{"unrecognized", "", false, "."},
		// Single dot provided for application name, so expect current
		// directory.
		{"windows", ".", false, "."},
		{"windows", ".", true, "."},
		{"linux", ".", false, "."},
		{"darwin", ".", false, "."},
		{"openbsd", ".", false, "."},
		{"freebsd", ".", false, "."},
		{"netbsd", ".", false, "."},
		{"plan9", ".", false, "."},
		{"unrecognized", ".", false, "."},
	}
	//t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		ret := TstAppDataDir(test.goos, test.appName, test.roaming)
		if ret != test.want {
			t.Errorf(
				"AppDataDir #%d (%s) does not match - "+
					"expected got %s, want %s", i, test.goos, ret,
				test.want,
			)
			continue
		}
	}
}

// TstAppDataDir makes the internal appDataDir function available to the test package.
func TstAppDataDir(goos, appName string, roaming bool) string {
	return appdata.GetDataDir(goos, appName, roaming)
}
