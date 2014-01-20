// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"launchpad.net/loggo"

	"launchpad.net/juju-core/juju/osenv"
)

var (
	aptLogger  = loggo.GetLogger("juju.utils.apt")
	aptProxyRE = regexp.MustCompile(`(?im)^\s*Acquire::(?P<protocol>[a-z]+)::Proxy\s+"(?P<proxy>[^"]+)";\s*$`)
)

// Some helpful functions for running apt in a sane way

// AptCommandOutput calls cmd.Output, this is used as an overloading point so we
// can test what *would* be run without actually executing another program
var AptCommandOutput = (*exec.Cmd).CombinedOutput

// This is the default apt-get command used in cloud-init, the various settings
// mean that apt won't actually block waiting for a prompt from the user.
var aptGetCommand = []string{
	"apt-get", "--option=Dpkg::Options::=--force-confold",
	"--option=Dpkg::options::=--force-unsafe-io", "--assume-yes", "--quiet",
}

// aptEnvOptions are options we need to pass to apt-get to not have it prompt
// the user
var aptGetEnvOptions = []string{"DEBIAN_FRONTEND=noninteractive"}

// AptGetInstall runs 'apt-get install packages' for the packages listed here
func AptGetInstall(packages ...string) error {
	cmdArgs := append([]string(nil), aptGetCommand...)
	cmdArgs = append(cmdArgs, "install")
	cmdArgs = append(cmdArgs, packages...)
	aptLogger.Infof("Running: %s", cmdArgs)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Env = append(os.Environ(), aptGetEnvOptions...)
	out, err := AptCommandOutput(cmd)
	if err != nil {
		aptLogger.Errorf("apt-get command failed: %v\nargs: %#v\n%s",
			err, cmdArgs, string(out))
		return fmt.Errorf("apt-get failed: %v", err)
	}
	return nil
}

// AptConfigProxy will consult apt-config about the configured proxy
// settings. If there are no proxy settings configured, an empty string is
// returned.
func AptConfigProxy() (string, error) {
	cmdArgs := []string{
		"apt-config",
		"dump",
		"Acquire::http::Proxy",
		"Acquire::https::Proxy",
		"Acquire::ftp::Proxy",
	}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	out, err := AptCommandOutput(cmd)
	if err != nil {
		aptLogger.Errorf("apt-config command failed: %v\nargs: %#v\n%s",
			err, cmdArgs, string(out))
		return "", fmt.Errorf("apt-config failed: %v", err)
	}
	return string(bytes.Join(aptProxyRE.FindAll(out, -1), []byte("\n"))), nil
}

// DetectAptProxies will shell out to apt-config to dump the http, https, and
// ftp proxy settings.
func DetectAptProxies() osenv.ProxySettings {
	return osenv.ProxySettings{}
}

// IsUbuntu executes lxb_release to see if the host OS is Ubuntu.
func IsUbuntu() bool {
	out, err := RunCommand("lsb_release", "-i", "-s")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) == "Ubuntu"
}

// IsPackageInstalled uses dpkg-query to determine if the `packageName`
// package is installed.
func IsPackageInstalled(packageName string) bool {
	_, err := RunCommand("dpkg-query", "--status", packageName)
	return err == nil
}
