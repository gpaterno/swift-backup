/*
 * swift-backup - A simple tool for backing up files to OpenStack Swift
 * (c) 2017 Giuseppe Paterno' <gpaterno@gpaterno.com>
 * Released under GPL v3 (see LICENSE)
 */

package main

import (
	"io"
	"net/http"
	"os"

	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"flag"
	"path"

	"strings"

	"github.com/alexcesaro/log/stdlog"
	"github.com/ncw/swift"
)

// Global vars
var (
	username   string
	password   string
	project    string
	authurl    string
	region     string
	deleteFile bool
)

// Init with the paramenters
func init() {
	flag.StringVar(&username, "os-username", "admin", "Swift username")
	flag.StringVar(&password, "os-password", "password", "Swift password")
	flag.StringVar(&project, "os-project", "admin", "Project name")
	flag.StringVar(&authurl, "os-authurl", "https://os.ch-ti1.server.one/v2.0/", "Keystone endpoint")
	flag.StringVar(&region, "os-region", "ch-ti1", "Region name")
	flag.BoolVar(&deleteFile, "delete-after", false, "Delete source file afterwards")
}

// There's no builtin to have a string in a slice
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Main (upload) function
func main() {

	logger := stdlog.GetFromFlags()
	flag.Parse()

	// Check if args
	if (len(flag.Args()) < 2) || (len(flag.Args()) > 2) {
		logger.Error("Usage: swift-backup <container> <file>")
		os.Exit(1)
	}

	// We are running semi-trusted cert, set to insecure
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConnsPerHost: 2048,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}

	// Setup container/filename vars
	var container = flag.Args()[0]
	var filename = flag.Args()[1]
	var tgt_filename = path.Base(filename)

	//logger.Debugf("Password: %s", password)
	logger.Debugf("Container: %s", container)
	logger.Debugf("Keystone endpoint: %s", authurl)
	logger.Debugf("Username: %s", username)
	logger.Debugf("Password: %s", strings.Repeat("*", len(password)))
	logger.Debugf("Project: %s", project)
	logger.Debugf("Filename: %s", filename)
	logger.Debugf("Target filename: %s", tgt_filename)

	c := swift.Connection{
		UserName:  username,
		ApiKey:    password,
		AuthUrl:   authurl,
		Region:    region,
		Tenant:    project,
		Transport: transport,
	}

	// Authenticate
	err := c.Authenticate()

	if err != nil {
		// Bad Request is a trap :)
		if err == swift.BadRequest {
			logger.Error("Authentication denied")
		} else {
			logger.Error(err)
		}
		os.Exit(1)
	}

	if !c.Authenticated() {
		logger.Error("Not authenticated")
		os.Exit(1)
	}

	// Check if container is among the list of containers
	containers, _ := c.ContainerNames(nil)
	logger.Debugf("Containers: %s", containers)

	if !stringInSlice(container, containers) {
		logger.Errorf("Container %s do not exist", container)
		os.Exit(1)
	}

	// Upload the file
	hash := md5.New()

	in, err := os.Open(filename)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	// Get hash of file
	io.Copy(hash, in)
	hashResult := hex.EncodeToString(hash.Sum(nil))
	logger.Debugf("Hash: %s", hashResult)
	in.Close()

	// Reopen to rewind
	in, err = os.Open(filename)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	// Create object in Swift
	out, err := c.ObjectCreate(container, tgt_filename, true, hashResult, "", nil)
	if err != nil {
		logger.Error(err)
	}

	// Copy the file to object storage
	if n, err := io.Copy(out, in); err != nil {
		logger.Error(err)
	} else {
		logger.Debugf("Copied %v bytes", n)
	}

	// Closing swift file and check for errors
	err = out.Close()

	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	// Closing input file
	in.Close()

	// Unlink if delete after
	// Check hash before deleting
	if deleteFile {
		obj, _, err := c.Object(container, filename)
		if err != nil {
			logger.Error("An error occurred while verifying hash")
			os.Exit(1)
		}

		logger.Debugf("Remote hash: %s", obj.Hash)
		logger.Debugf("Local hash:  %s", hashResult)

		// If hashes do not match, do not delete, corruption happened :(
		if obj.Hash != hashResult {
			logger.Error("Hash error while comparing local and remote, cowardly refusing removing local file")
			os.Exit(1)
		}

		// Now deleting file
		logger.Debug("Hashes matches, deleting file")
		err = os.Remove(filename)

		if err != nil {
			logger.Error("An error occurred while removing source file")
			os.Exit(1)
		}
	}

	// Clean exit :)
	os.Exit(0)

}
