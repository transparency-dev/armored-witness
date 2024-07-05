package devices

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"strings"

	"golang.org/x/mod/sumdb/note"
)

const (
	AttestWitnessID = "ArmoredWitness ID attestation v1"
	AttestBastionID = "ArmoredWitness BastionID attestation v1"
)

var (
	//go:embed ci/*
	ci embed.FS
	//go:embed prod/*
	prod embed.FS

	CI   map[string]Device
	Prod map[string]Device
)

func init() {
	var err error
	if CI, err = parseFS(ci); err != nil {
		panic(err)
	}
	if Prod, err = parseFS(prod); err != nil {
		panic(err)
	}
}

// Device represents an ArmoredWitness device and its various attested identities.
type Device struct {
	ID            string
	BastionID     string
	WitnessPubkey string
}

type entry struct {
	pubKey       string
	attestations [][]byte
}

func parseFS(f embed.FS) (map[string]Device, error) {
	entries := make(map[string]entry)
	errs := []error{}
	err := fs.WalkDir(f, ".", func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}

		n := d.Name()
		parts := strings.SplitN(n, ".", 2)
		if len(parts) != 2 {
			errs = append(errs, fmt.Errorf("badly named file %q", n))
			return nil
		}
		body, err := f.ReadFile(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read %q: %v", n, err))
			return nil
		}
		e := entries[parts[0]]
		if parts[1] == "pub" {
			e.pubKey = string(body)

		} else {
			e.attestations = append(e.attestations, body)
		}

		entries[parts[0]] = e
		return nil
	})
	if err != nil {
		return nil, err
	}

	r := make(map[string]Device)
	for _, v := range entries {
		d, err := new(v.pubKey, v.attestations)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		r[d.ID] = *d
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return r, nil
}

func new(attestPub string, attestations [][]byte) (*Device, error) {
	v, err := note.NewVerifier(attestPub)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", attestPub, err)
	}

	d := &Device{ID: v.Name()}

	for _, a := range attestations {
		n, err := note.Open(a, note.VerifierList(v))
		if err != nil {
			return nil, fmt.Errorf("couldn't open %q: %v", a, err)

		}
		lines := strings.Split(n.Text, "\n")
		switch lines[0] {
		case AttestWitnessID:
			if len(lines) < 4 {
				return nil, fmt.Errorf("%s: invalid ID attestation", v.Name())
			}
			d.WitnessPubkey = lines[3]
		case AttestBastionID:
			if len(lines) < 4 {
				log.Printf("%q", n.Text)
				return nil, fmt.Errorf("%s: invalid bastion attestation (%d)", v.Name(), len(lines))
			}
			d.BastionID = lines[3]
		default:
			continue
		}
	}
	return d, nil
}
