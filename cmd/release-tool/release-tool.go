package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SkycoinProject/skycoin/src/cipher"
	"github.com/SkycoinProject/skycoin/src/cipher/encoder"
)

// Build represents the build file info
type Build struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
	Size int64  `json:"size"`
}

// Release represents the release info
type Release struct {
	Version string  `json:"version"`
	Date    string  `json:"date"`
	Builds  []Build `json:"build"`
	Count   int64   `json:"count"` // count of builds
}

// Manifest file struct
type Manifest struct {
	Release     Release `json:"release"`
	Note        string  `json:"note"`
	ReleaseHash string  `json:"release_hash"`
	Pubkey      string  `json:"pubkey"`
	Sig         string  `json:"signature"`
}

var preUsage = `
Skycoin release sign and verify tool
Version: v0.1.0

Syntax: %s [options] [builds dir | manifest file]

Example:
# Create and sign the release manifest
%s -sign -seckey $SECRET_KEY -note "release v0.26.0" electron/release/ > manifest.json

# Verify the manifest file
%s -verify -pubkey $PUBLIC_KEY manifest.json

`

func main() {
	sign := flag.Bool("sign", false, "sign the builds and generate a release manifest")
	verify := flag.Bool("verify", false, "verify the manifest")
	pubkeyStr := flag.String("pubkey", "", "pubkey for verifying the release")
	seckeyStr := flag.String("seckey", "", "secret key for signing the release")
	note := flag.String("note", "", "note as salt, could be the sha256 of the last manifest file")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), preUsage, os.Args[0], os.Args[0], os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Commands:\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// geneate sha256 hash of each build and sign the hashes.
	if *sign {
		if *seckeyStr == "" {
			fmt.Fprintf(os.Stderr, "seckey is required for signing, set it with -seckey flag")
			return
		}

		if *note == "" {
			fmt.Fprintln(os.Stderr, "note is missing, set it with -note flag")
			return
		}

		sk, err := cipher.SecKeyFromHex(*seckeyStr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid seckey:", err)
			return
		}

		dir := flag.Arg(0)
		if dir == "" {
			fmt.Fprintf(os.Stderr, "build directory is missing. \nRun '%s -h' to see the usage.\n", os.Args[0])
			return
		}

		v, err := hashAndSignBuilds(dir, sk, *note)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		fmt.Println(string(v))
		return
	}

	if *verify {
		if *pubkeyStr == "" {
			fmt.Fprintln(os.Stderr, "pubkey is required for verifying, set it with -pubkey flag")
			return
		}
		pk, err := cipher.PubKeyFromHex(*pubkeyStr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid pubkey:", err)
			return
		}

		input := flag.Arg(0)
		if input == "" {
			fmt.Fprintf(os.Stderr, "manifest file is missing. \nRun '%s -h' to see the usage.\n", os.Args[0])
			return
		}

		if err := verifyManifest(input, pk); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		fmt.Println("Verify success")
		return
	}

	flag.Usage()
}

func hashAndSignBuilds(dir string, sk cipher.SecKey, note string) ([]byte, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("No build files in dir: %s", dir)
	}

	builds := make([]Build, 0, len(files))
	var version string
	for _, f := range files {
		// Ignore subdirectory
		if f.IsDir() {
			continue
		}

		// Ignore invisable file. e.g: .DS_Store
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}

		// the build file names should start with skycoin-$VERSION
		if !strings.HasPrefix(f.Name(), "skycoin") {
			continue
		}

		ss := strings.Split(f.Name(), "-")
		// Compare the versions, the version should all be the same, otherwise return with error
		if version != "" && version != ss[1] {
			return nil, fmt.Errorf("Different build versions of %s and %s exist in dir: %s", version, ss[1], dir)
		}

		version = ss[1]

		d, err := ioutil.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			return nil, err
		}
		hash := cipher.SumSHA256(d)
		builds = append(builds, Build{
			Name: f.Name(),
			Size: f.Size(),
			Hash: hash.Hex(),
		})
	}

	if len(builds) == 0 {
		return nil, fmt.Errorf("There's no build file of version %s in dir: %s", version, dir)
	}

	t := time.Now()
	date := fmt.Sprintf("%d-%d-%d", t.Year(), t.Month(), t.Day())

	pk, err := cipher.PubKeyFromSecKey(sk)
	if err != nil {
		return nil, fmt.Errorf("generate pubkey from seckey failed: %v", err)
	}

	mf := Manifest{
		Release: Release{
			Version: version,
			Date:    date,
			Builds:  builds,
			Count:   int64(len(builds)),
		},
		Note:   note,
		Pubkey: pk.Hex(),
	}

	rh := cipher.SumSHA256(encoder.Serialize(&mf.Release))

	mf.ReleaseHash = rh.Hex()
	saltHash := cipher.SumSHA256([]byte(mf.Note))
	// SHA256(hash of release field, hash of note as salt)
	hash2Sign := cipher.AddSHA256(rh, saltHash)
	mf.Sig = cipher.MustSignHash(hash2Sign, sk).Hex()

	return json.MarshalIndent(mf, "  ", "  ")
}

func verifyManifest(manifest string, pubkey cipher.PubKey) error {
	mf := Manifest{}
	d, err := ioutil.ReadFile(manifest)
	if err != nil {
		return err
	}
	if err := json.NewDecoder(bytes.NewReader(d)).Decode(&mf); err != nil {
		return err
	}

	ch := cipher.SumSHA256(encoder.Serialize(mf.Release))
	if mf.ReleaseHash != ch.Hex() {
		return fmt.Errorf("Release hash does not match")
	}

	if mf.Pubkey != pubkey.Hex() {
		return fmt.Errorf("Pubkeys do not match")
	}

	if mf.Sig == "" {
		return fmt.Errorf("Signature filed in manifest is missing")
	}

	sig, err := cipher.SigFromHex(mf.Sig)
	if err != nil {
		return err
	}

	saltHash := cipher.SumSHA256([]byte(mf.Note))

	hash2Sign := cipher.AddSHA256(ch, saltHash)

	if err := cipher.VerifySignatureRecoverPubKey(sig, hash2Sign); err != nil {
		return err
	}

	if err := cipher.VerifyPubKeySignedHash(pubkey, sig, hash2Sign); err != nil {
		return err
	}

	return nil
}
