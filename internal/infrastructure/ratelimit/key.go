package ratelimit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type KeyScope string

const (
	ScopeIP         KeyScope = "ip"
	ScopeUser       KeyScope = "user"
	ScopeIdentifier KeyScope = "identifier"
	ScopeEmail      KeyScope = "email"
)

type KeyBuilder struct {
	Prefix string
}

func NewKeyBuilder(prefix string) KeyBuilder {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "portal:rl"
	}

	return KeyBuilder{Prefix: prefix}
}

func (b KeyBuilder) Build(policyName string, scope KeyScope, value string) string {
	policyName = normalize(policyName)
	value = normalize(value)

	if value == "" {
		value = "unknown"
	}

	return fmt.Sprintf("%s:%s:%s:%s", b.Prefix, policyName, scope, hash(value))
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])[:32]
}
