package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	ecr "github.com/awslabs/amazon-ecr-credential-helper/ecr-login"
	"github.com/chrismellard/docker-credential-acr-env/pkg/credhelper"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/github"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"

	"github.com/docker-credential-magic/docker-credential-magic/internal/constants"
)

var (
	// Version can be set via:
	// -ldflags="-X main.Version=$TAG"
	Version string
)

func main() {
	args := os.Args
	if len(args) < 2 {
		usage()
	}
	subcommand := args[1]
	switch subcommand {
	case constants.HelperSubcommandGet:
		subcommandGet()
	case "version":
		subcommandVersion()
	default:
		usage()
		log.Fatalf("not implemented:", subcommand)
	}
}

func usage() {
	fmt.Printf("Usage: docker-credential-magic <%s|home|init|version>\n",
		constants.HelperSubcommandGet)
	os.Exit(1)
}

// keychain provides cred helper logic for common registry implementations.
var keychain = authn.NewMultiKeychain(
	google.Keychain,
	github.Keychain,
	authn.NewKeychainFromHelper(ecr.NewECRHelper(ecr.WithLogger(ioutil.Discard))),
	authn.NewKeychainFromHelper(credhelper.NewACRCredentialsHelper()),
)

func subcommandGet() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	rawInput := scanner.Text()
	reg, err := name.NewRegistry(rawInput)
	if err != nil {
		log.Fatalf("parsing registry server URL: %v", err)
	}

	if auth, _ := keychain.Resolve(reg); auth != authn.Anonymous {
		ac, err := auth.Authorization()
		if err != nil {
			log.Fatalf("getting creds: %v", err)
		}
		if err := json.NewEncoder(os.Stdout).Encode(toCreds(ac)); err != nil {
			log.Fatal("[magic] json.Encode: %v", err)
		}
	}
	log.Fatalf("no matching credentials for %q", rawInput)
}

func subcommandVersion() {
	fmt.Println(Version)
	os.Exit(0)
}

// Borrowed from:
// https://github.com/google/go-containerregistry/blob/a0b9468898deb31e3eb35f97fa4f0d568e769296/cmd/crane/cmd/auth.go#L47
type credentials struct {
	Username string
	Secret   string
}

// Borrowed from:
// https://github.com/google/go-containerregistry/blob/a0b9468898deb31e3eb35f97fa4f0d568e769296/cmd/crane/cmd/auth.go#L53
func toCreds(config *authn.AuthConfig) credentials {
	creds := credentials{
		Username: config.Username,
		Secret:   config.Password,
	}

	if config.IdentityToken != "" {
		creds.Username = "<token>"
		creds.Secret = config.IdentityToken
	}
	return creds
}
