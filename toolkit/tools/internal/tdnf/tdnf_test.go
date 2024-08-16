// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tdnf

import (
	"os"
	"testing"

	"github.com/microsoft/azurelinux/toolkit/tools/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logger.InitStderrLog()
	os.Exit(m.Run())
}

func TestGetMajorVersionFromString_AcceptEmptyMinorVersion(t *testing.T) {
	fullVersion := "2.0"
	expectedMajorVersion := "2.0"
	majorVersion, err := getMajorVersionFromString(fullVersion)

	assert.NoError(t, err)
	assert.Equal(t, expectedMajorVersion, majorVersion)
}

func TestGetMajorVersionFromString_AcceptPopulatedMinorVersionTimestamp(t *testing.T) {
	fullVersion := "2.0.20220519.1844"
	expectedMajorVersion := "2.0"
	majorVersion, err := getMajorVersionFromString(fullVersion)

	assert.NoError(t, err)
	assert.Equal(t, expectedMajorVersion, majorVersion)
}

func TestGetMajorVersionFromString_AcceptAlphabeticalMinorVersion(t *testing.T) {
	fullVersion := "2.0.reallycoolteststring"
	expectedMajorVersion := "2.0"
	majorVersion, err := getMajorVersionFromString(fullVersion)

	assert.NoError(t, err)
	assert.Equal(t, expectedMajorVersion, majorVersion)
}

func TestGetMajorVersionFromString_RejectAlphabeticalInMajor(t *testing.T) {
	fullVersion := "2.A"
	_, err := getMajorVersionFromString(fullVersion)
	assert.Error(t, err)
}

func TestGetMajorVersionFromString_RejectEmpty(t *testing.T) {
	fullVersion := ""
	_, err := getMajorVersionFromString(fullVersion)
	assert.Error(t, err)
}

func TestGetMajorVersionFromString_RejectWithNoDot(t *testing.T) {
	fullVersion := "2"
	_, err := getMajorVersionFromString(fullVersion)
	assert.Error(t, err)
}

func TestGetMajorVersionFromString_RejectNoSecondComponentWithDot(t *testing.T) {
	fullVersion := "2."
	_, err := getMajorVersionFromString(fullVersion)
	assert.Error(t, err)
}

func TestGetMajorVersionFromString_RejectTrailingDot(t *testing.T) {
	fullVersion := "2.0."
	_, err := getMajorVersionFromString(fullVersion)
	assert.Error(t, err)
}

func TestInstallPackageRegex_MatchesPackageName(t *testing.T) {
	const line = "X aarch64 1.1b.8_X-22~rc1.azl3 fetcher-cloned-repo"

	matches := InstallPackageRegex.FindStringSubmatch(line)

	assert.Len(t, matches, InstallMaxMatchLen)
	assert.Equal(t, "X", matches[InstallPackageName])
}

func TestInstallPackageRegex_FailsForMissingPackageName(t *testing.T) {
	const line = " aarch64 1.1b.8_X-22~rc1.azl3 fetcher-cloned-repo"

	assert.False(t, InstallPackageRegex.MatchString(line))
}

func TestInstallPackageRegex_MatchesPackageArch(t *testing.T) {
	const line = "X aarch64 1.1b.8_X-22~rc1.azl3 fetcher-cloned-repo"

	matches := InstallPackageRegex.FindStringSubmatch(line)

	assert.Len(t, matches, InstallMaxMatchLen)
	assert.Equal(t, "aarch64", matches[InstallPackageArch])
}

func TestInstallPackageRegex_FailsForMissingArch(t *testing.T) {
	const line = "X  1.1b.8_X-22~rc1.azl3 fetcher-cloned-repo"

	assert.False(t, InstallPackageRegex.MatchString(line))
}

func TestInstallPackageRegex_MatchesPackageVersionNoEpoch(t *testing.T) {
	const line = "X aarch64 1.1b.8_X-22~rc1.azl3 fetcher-cloned-repo"

	matches := InstallPackageRegex.FindStringSubmatch(line)

	assert.Len(t, matches, InstallMaxMatchLen)
	assert.Equal(t, "1.1b.8_X-22~rc1", matches[InstallPackageVersion])
}

func TestInstallPackageRegex_MatchesPackageVersionWithEpoch(t *testing.T) {
	const line = "X aarch64 5:1.1b.8_X-22~rc1.azl3 fetcher-cloned-repo"

	matches := InstallPackageRegex.FindStringSubmatch(line)

	assert.Len(t, matches, InstallMaxMatchLen)
	assert.Equal(t, "5:1.1b.8_X-22~rc1", matches[InstallPackageVersion])
}

func TestInstallPackageRegex_FailsForMissingVersion(t *testing.T) {
	const line = "X aarch64 .azl3 fetcher-cloned-repo"

	assert.False(t, InstallPackageRegex.MatchString(line))
}

func TestInstallPackageRegex_MatchesPackageDist(t *testing.T) {
	const line = "X aarch64 1.1b.8_X-22~rc1.azl3 fetcher-cloned-repo"

	matches := InstallPackageRegex.FindStringSubmatch(line)

	assert.Len(t, matches, InstallMaxMatchLen)
	assert.Equal(t, "azl3", matches[InstallPackageDist])
}

func TestInstallPackageRegex_FailsForMissingDist(t *testing.T) {
	const line = "X aarch64 1.1b.8_X-22~rc1 fetcher-cloned-repo"

	assert.False(t, InstallPackageRegex.MatchString(line))
}

func TestInstallPackageRegex_MatchesRandomWhiteSpaces(t *testing.T) {
	const line = "X   aarch64  1.1b.8_X-22~rc1.azl3          	fetcher-cloned-repo"

	assert.True(t, InstallPackageRegex.MatchString(line))
}

func TestInstallPackageRegex_DoesNotMatchInvalidLine(t *testing.T) {
	const line = "Invalid line"

	assert.False(t, InstallPackageRegex.MatchString(line))
}

func Test_getRepoSnapshotCliArg(t *testing.T) {
	type args struct {
		posixTime string
	}
	tests := []struct {
		name             string
		args             args
		wantRepoSnapshot string
		wantErr          bool
	}{
		{name: "testEmpty", args: args{posixTime: ""}, wantRepoSnapshot: "", wantErr: true},
		{name: "testIsNotNumeric", args: args{posixTime: "12345qwerty"}, wantRepoSnapshot: "", wantErr: true},
		{name: "testNumeric", args: args{posixTime: "123456789"}, wantRepoSnapshot: "--snapshottime=123456789", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRepoSnapshot, err := getRepoSnapshotCliArg(tt.args.posixTime)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantRepoSnapshot, gotRepoSnapshot)
		})
	}
}

func Test_getRepoSnapshotExcludeCliArg(t *testing.T) {
	type args struct {
		excludeRepo string
	}
	tests := []struct {
		name           string
		args           args
		wantExcludeArg string
		wantErr        bool
	}{
		{name: "testEmpty", args: args{excludeRepo: ""}, wantExcludeArg: "", wantErr: true},
		{name: "testRepo", args: args{excludeRepo: "local-repo"}, wantExcludeArg: "--excludesnapshot=local-repo", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExcludeArg, err := getRepoSnapshotExcludeCliArg(tt.args.excludeRepo)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRepoSnapshotExcludeCliArg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotExcludeArg != tt.wantExcludeArg {
				t.Errorf("getRepoSnapshotExcludeCliArg() = %v, want %v", gotExcludeArg, tt.wantExcludeArg)
			}
		})
	}
}
