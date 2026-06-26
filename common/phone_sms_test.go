package common

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizePhoneNumber(t *testing.T) {
	require.Equal(t, "+8613800138000", NormalizePhoneNumber(" 138-0013-8000 ", "86"))
	require.Equal(t, "+15551234567", NormalizePhoneNumber("+1 (555) 123-4567", "86"))
	require.Equal(t, "+15551234567", NormalizePhoneNumber("0015551234567", "86"))
	require.Equal(t, "", NormalizePhoneNumber("not-a-phone", "86"))
}

func TestResolveSMSIHuyiSettingsPrefersEnvOverOptionMap(t *testing.T) {
	OptionMapRWMutex.Lock()
	originalMap := OptionMap
	OptionMap = map[string]string{
		"SMSIHuyiEnabled":    "false",
		"SMSIHuyiAPIID":      "db-account",
		"SMSIHuyiAPIKey":     "db-key",
		"SMSIHuyiTemplateID": "db-template",
	}
	OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		OptionMapRWMutex.Lock()
		OptionMap = originalMap
		OptionMapRWMutex.Unlock()
	})

	t.Setenv("SMS_IHUYI_ENABLED", "true")
	t.Setenv("SMS_IHUYI_API_ID", "env-account")
	t.Setenv("SMS_IHUYI_API_KEY", "env-key")
	t.Setenv("SMS_IHUYI_TEMPLATE_ID", "env-template")

	settings := ResolveSMSIHuyiSettings()
	require.True(t, settings.Enabled)
	require.Equal(t, "env-account", settings.Account)
	require.Equal(t, "env-key", settings.Password)
	require.Equal(t, "env-template", settings.TemplateID)
}

func TestResolveSMSIHuyiSettingsFallsBackToOptionMap(t *testing.T) {
	for _, key := range []string{
		"SMS_IHUYI_ENABLED",
		"SMS_IHUYI_API_ID",
		"SMS_IHUYI_API_KEY",
		"SMS_IHUYI_TEMPLATE_ID",
	} {
		require.NoError(t, os.Unsetenv(key))
	}

	OptionMapRWMutex.Lock()
	originalMap := OptionMap
	OptionMap = map[string]string{
		"SMSIHuyiEnabled":    "true",
		"SMSIHuyiAPIID":      "db-account",
		"SMSIHuyiAPIKey":     "db-key",
		"SMSIHuyiTemplateID": "db-template",
	}
	OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		OptionMapRWMutex.Lock()
		OptionMap = originalMap
		OptionMapRWMutex.Unlock()
	})

	settings := ResolveSMSIHuyiSettings()
	require.True(t, settings.Enabled)
	require.Equal(t, "db-account", settings.Account)
	require.Equal(t, "db-key", settings.Password)
	require.Equal(t, "db-template", settings.TemplateID)
}

func TestResolveSMSIHuyiSettingsDefaultsEnabled(t *testing.T) {
	OptionMapRWMutex.Lock()
	originalMap := OptionMap
	OptionMap = nil
	OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		OptionMapRWMutex.Lock()
		OptionMap = originalMap
		OptionMapRWMutex.Unlock()
	})

	for _, key := range []string{
		"SMS_IHUYI_ENABLED",
		"SMS_IHUYI_API_ID",
		"SMS_IHUYI_API_KEY",
		"SMS_IHUYI_TEMPLATE_ID",
	} {
		require.NoError(t, os.Unsetenv(key))
	}

	settings := ResolveSMSIHuyiSettings()
	require.True(t, settings.Enabled)
	require.Equal(t, defaultIHuyiSMSTemplateID, settings.TemplateID)
}

func TestIsPhoneVerificationEnabledPrefersEnv(t *testing.T) {
	OptionMapRWMutex.Lock()
	originalMap := OptionMap
	originalEnabled := PhoneVerificationEnabled
	OptionMap = map[string]string{"PhoneVerificationEnabled": "false"}
	PhoneVerificationEnabled = false
	OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		OptionMapRWMutex.Lock()
		OptionMap = originalMap
		PhoneVerificationEnabled = originalEnabled
		OptionMapRWMutex.Unlock()
	})

	t.Setenv("PHONE_VERIFICATION_ENABLED", "true")

	require.True(t, IsPhoneVerificationEnabled())
}

func TestIsPhoneVerificationEnabledFallsBackToOptionMap(t *testing.T) {
	OptionMapRWMutex.Lock()
	originalMap := OptionMap
	originalEnabled := PhoneVerificationEnabled
	OptionMap = map[string]string{"PhoneVerificationEnabled": "true"}
	PhoneVerificationEnabled = false
	OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		OptionMapRWMutex.Lock()
		OptionMap = originalMap
		PhoneVerificationEnabled = originalEnabled
		OptionMapRWMutex.Unlock()
	})

	require.NoError(t, os.Unsetenv("PHONE_VERIFICATION_ENABLED"))

	require.True(t, IsPhoneVerificationEnabled())
}

func TestIsPhoneVerificationEnabledDefaultsEnabled(t *testing.T) {
	OptionMapRWMutex.Lock()
	originalMap := OptionMap
	originalEnabled := PhoneVerificationEnabled
	OptionMap = nil
	PhoneVerificationEnabled = true
	OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		OptionMapRWMutex.Lock()
		OptionMap = originalMap
		PhoneVerificationEnabled = originalEnabled
		OptionMapRWMutex.Unlock()
	})

	require.NoError(t, os.Unsetenv("PHONE_VERIFICATION_ENABLED"))

	require.True(t, IsPhoneVerificationEnabled())
}
