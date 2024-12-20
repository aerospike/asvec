//go:build unit

package flags

import (
	"testing"

	"github.com/aerospike/asvec/utils"
)

func TestCredentialsFlag_Set(t *testing.T) {
	// Test setting user and password
	flag := CredentialsFlag{}
	err := flag.Set("username:password")
	if err != nil {
		t.Errorf("Error setting credentials: %v", err)
	}

	// Test setting user only
	err = flag.Set("username")
	if err != nil {
		t.Errorf("Error setting user: %v", err)
	}

	// Test setting password only
	err = flag.Set(":password")
	if err != nil {
		t.Errorf("Error setting password: %v", err)
	}

	// Test setting empty value
	err = flag.Set("")
	if err != nil {
		t.Errorf("Error setting empty value: %v", err)
	}
}

func TestCredentialsFlag_Type(t *testing.T) {
	flag := CredentialsFlag{}
	expected := "<user>[:<pass>]"
	actual := flag.Type()

	if expected != actual {
		t.Errorf("Expected type '%s', got '%s'", expected, actual)
	}
}

func TestCredentialsFlag_String(t *testing.T) {
	// Test string representation with user and password
	flag := CredentialsFlag{
		User:     StringOptionalFlag{Val: utils.Ptr("username")},
		Password: StringOptionalFlag{Val: utils.Ptr("password")},
	}
	str := flag.String()
	expected := "username:password"
	if str != expected {
		t.Errorf("Expected string '%s', got '%s'", expected, str)
	}

	// Test string representation with user only
	flag = CredentialsFlag{
		User:     StringOptionalFlag{Val: utils.Ptr("username")},
		Password: StringOptionalFlag{},
	}
	str = flag.String()
	expected = "username:<nil>"
	if str != expected {
		t.Errorf("Expected string '%s', got '%s'", expected, str)
	}

	// Test string representation with empty values
	flag = CredentialsFlag{
		User:     StringOptionalFlag{},
		Password: StringOptionalFlag{},
	}
	str = flag.String()
	expected = "<nil>:<nil>"
	if str != expected {
		t.Errorf("Expected string '%s', got '%s'", expected, str)
	}
}
