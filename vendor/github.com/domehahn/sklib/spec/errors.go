package spec

import "errors"

// Sentinel errors for common validation failures.
var (
	ErrEmptyName        = errors.New("name must not be empty")
	ErrInvalidName      = errors.New("invalid skill name: must match ^[a-z0-9]+(?:-[a-z0-9]+)*$")
	ErrEmptyNamespace   = errors.New("namespace must not be empty")
	ErrInvalidNamespace = errors.New("invalid namespace: must match ^[a-z0-9]+(?:[._-][a-z0-9]+)*$")
	ErrEmptyVersion     = errors.New("version must not be empty")
	ErrInvalidVersion   = errors.New("version is not valid SemVer")
	ErrEmptyDescription = errors.New("description must not be empty")
	ErrNoPlatforms      = errors.New("compatible_with must not be empty")
	ErrEmptyEntrypoint  = errors.New("entrypoint must not be empty")
)
