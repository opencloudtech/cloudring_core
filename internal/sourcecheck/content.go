// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package sourcecheck

import (
	"net/netip"
	"regexp"
	"strings"
	"unicode"
)

type contentRule struct {
	id      string
	class   string
	message string
	pattern *regexp.Regexp
}

var directContentRules = []contentRule{
	{id: "github_classic_token", class: "secret", message: "GitHub token-like value", pattern: regexp.MustCompile(`\b` + "g" + `hp_[A-Za-z0-9_]{20,}\b`)},
	{id: "github_fine_grained_token", class: "secret", message: "GitHub fine-grained token-like value", pattern: regexp.MustCompile(`\b` + "github_" + `pat_[A-Za-z0-9_]{20,}\b`)},
	{id: "cloud_access_key", class: "secret", message: "cloud access-key identifier", pattern: regexp.MustCompile(`\b(?:A` + `KIA|A` + `SIA)[A-Z0-9]{16}\b`)},
	{id: "private_key_block", class: "private_key", message: "private key material", pattern: regexp.MustCompile(`B` + `EGIN [A-Z0-9 ]*PRIVATE KEY`)},
	{id: "authorization_bearer", class: "secret", message: "serialized authorization credential", pattern: regexp.MustCompile(`(?i)Authori` + `zation\s*:\s*Bearer\s+[-._~+/A-Za-z0-9]+=*`)},
	{id: "compact_jwt", class: "secret", message: "compact JWT-like credential", pattern: regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b`)},
	{id: "kubeconfig_key_data", class: "kubeconfig", message: "embedded kubeconfig credential material", pattern: regexp.MustCompile(`(?i)client-` + `key-data\s*:`)},
	{id: "kubeconfig_certificate_data", class: "kubeconfig", message: "embedded kubeconfig credential material", pattern: regexp.MustCompile(`(?i)(?:client-` + `certificate-data|certificate-` + `authority-data)\s*:`)},
	{id: "kubeconfig_context", class: "kubeconfig", message: "serialized kubeconfig context", pattern: regexp.MustCompile(`(?i)current-` + `context\s*:`)},
}

var (
	credentialKeyPattern        = regexp.MustCompile(`(?i)(?:^|[\s,{])["']?([a-z][a-z0-9_-]*)["']?\s*([:=])`)
	environmentReferencePattern = regexp.MustCompile(`^(?:\$\{[A-Z_][A-Z0-9_]*\}|\$[A-Z_][A-Z0-9_]*|(?:os[.])?(?:Getenv|LookupEnv)[(]["'][A-Z_][A-Z0-9_]*["'][)])$`)
	environmentNamePattern      = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)
	dnsReferencePattern         = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9.-]{0,251}[a-z0-9])?$`)
	structuredReferencePattern  = regexp.MustCompile(`^(?:secretref|secretstore|credentialref)[.][a-z0-9](?:[a-z0-9._:/-]*[a-z0-9])?$`)
	ipv4TokenPattern            = regexp.MustCompile(`\b(?:[0-9]{1,3}[.]){3}[0-9]{1,3}\b`)
	ipv6TokenPattern            = regexp.MustCompile(`\[?[0-9A-Fa-f:]*:[0-9A-Fa-f:]+\]?`)
	privateHostnamePattern      = regexp.MustCompile(`(?i)(?:https?://|ssh://|git@)(?:[a-z0-9-]+[.])+(?:in` + `ternal|lo` + `cal)(?::[0-9]+)?`)
	barePrivateHostnamePattern  = regexp.MustCompile(`(?i)\b(?:[a-z0-9-]+[.])+(?:in` + `ternal|lo` + `cal)(?::[0-9]+)?\b`)
	unixUserPathPattern         = regexp.MustCompile(`(?i)/(?:Us` + `ers|ho` + `me)/[^\s"']+`)
	windowsUserPathPattern      = regexp.MustCompile(`(?i)[a-z]:[\\/](?:Us` + `ers|Documents and Settings)[\\/][^\s"']+`)
	credentialKeys              = exactStringSet(
		"access"+"key", "access"+"keyid", "access"+"token", "api"+"key", "api"+"token", "auth"+"token",
		"client"+"secret", "credential", "pass"+"word", "pass"+"wd", "private"+"key", "refresh"+"token",
		"se"+"cret", "secret"+"value", "signing"+"key", "to"+"ken",
	)
	referenceCredentialKeys = exactStringSet(
		"access"+"keyref", "api"+"keyref", "client"+"secretref", "credential"+"ref", "pass"+"wordref",
		"private"+"keyref", "secret"+"keyref", "secret"+"name", "secret"+"ref", "secret"+"storeref", "token"+"env",
	)
)

func scanContent(path string, content string) []Finding {
	budget := newFindingBudget(nil)
	findings, _ := scanContentWithBudget(path, content, budget)
	return findings
}

func scanContentWithBudget(path string, content string, budget *findingBudget) ([]Finding, error) {
	var findings []Finding
	lineNumber := 1
	for start := 0; start <= len(content); {
		end := strings.IndexByte(content[start:], '\n')
		if end < 0 {
			end = len(content)
		} else {
			end += start
		}
		line := content[start:end]
		if err := budget.consumeLine(); err != nil {
			return nil, err
		}
		for _, rule := range directContentRules {
			if rule.pattern.FindStringIndex(line) != nil {
				if err := budget.add(&findings, Finding{Rule: rule.id, Class: rule.class, Line: lineNumber, Message: rule.message}); err != nil {
					return nil, err
				}
			}
		}
		if credentialAssignment(line) {
			if err := budget.add(&findings, Finding{Rule: "credential_assignment", Class: "secret", Line: lineNumber, Message: "credential-like assignment contains a value instead of a reference"}); err != nil {
				return nil, err
			}
		}
		if privateEndpoint(line) {
			if err := budget.add(&findings, Finding{Rule: "private_endpoint", Class: "private_endpoint", Line: lineNumber, Message: "private or host-local endpoint value"}); err != nil {
				return nil, err
			}
		}
		if localFilesystemPath(line) {
			if err := budget.add(&findings, Finding{Rule: "local_user_path", Class: "local_path", Line: lineNumber, Message: "local user-profile path"}); err != nil {
				return nil, err
			}
		}
		if privateTreeReference(line) {
			if err := budget.add(&findings, Finding{Rule: "private_tree_reference", Class: "private_source", Line: lineNumber, Message: "private source-tree or evidence reference"}); err != nil {
				return nil, err
			}
		}
		if privateSourceAttribution(line) {
			if err := budget.add(&findings, Finding{Rule: "private_source_attribution", Class: "private_source", Line: lineNumber, Message: "attribution to copied private or proprietary source"}); err != nil {
				return nil, err
			}
		}
		if end == len(content) {
			break
		}
		start = end + 1
		lineNumber++
	}
	readiness, err := readinessFindingsWithBudget(path, content, budget)
	if err != nil {
		return nil, err
	}
	return append(findings, readiness...), nil
}

func credentialAssignment(line string) bool {
	if strings.HasPrefix(strings.TrimSpace(line), "type ") {
		return false
	}
	for _, match := range credentialKeyPattern.FindAllStringSubmatchIndex(line, -1) {
		if len(match) != 6 {
			continue
		}
		separatorStart := match[4]
		valueStart := match[1]
		separator := line[separatorStart]
		if separator == '=' && valueStart < len(line) && line[valueStart] == '=' {
			continue
		}
		if separator == ':' && valueStart < len(line) && line[valueStart] == '=' {
			valueStart++
		}
		key := normalizeCredentialKey(line[match[2]:match[3]])
		value := credentialScalar(line[valueStart:])
		// An empty YAML sensitive-field mapping followed by indented structural
		// fields contains no credential scalar. Nested lines are scanned
		// independently, so treating the parent mapping as safe cannot hide an
		// assigned credential value.
		if value == "" {
			continue
		}
		if referenceCredentialKey(key) {
			if !validReferenceCredentialValue(key, value) {
				return true
			}
			continue
		}
		if !credentialKey(key) {
			continue
		}
		if !structuralCredentialReference(value) {
			return true
		}
	}
	return false
}

func credentialScalar(remainder string) string {
	value := strings.TrimLeft(remainder, " \t")
	if value == "" {
		return ""
	}
	parenDepth, braceDepth, bracketDepth := 0, 0, 0
	var quote byte
	escaped := false
	for index := 0; index < len(value); index++ {
		character := value[index]
		if quote != 0 {
			switch {
			case escaped:
				escaped = false
			case character == '\\':
				escaped = true
			case character == quote:
				quote = 0
			}
			continue
		}
		switch character {
		case '"', '\'':
			quote = character
		case '(':
			parenDepth++
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
		case '{':
			braceDepth++
		case '}':
			if braceDepth == 0 && parenDepth == 0 && bracketDepth == 0 {
				return strings.TrimSpace(value[:index])
			}
			braceDepth--
		case '[':
			bracketDepth++
		case ']':
			if bracketDepth == 0 && parenDepth == 0 && braceDepth == 0 {
				return strings.TrimSpace(value[:index])
			}
			bracketDepth--
		case ',':
			if parenDepth == 0 && braceDepth == 0 && bracketDepth == 0 {
				return strings.TrimSpace(value[:index])
			}
		case '#':
			if parenDepth == 0 && braceDepth == 0 && bracketDepth == 0 && (index == 0 || value[index-1] == ' ' || value[index-1] == '\t') {
				return strings.TrimSpace(value[:index])
			}
		}
	}
	return strings.TrimSpace(value)
}

func normalizeCredentialKey(value string) string {
	var builder strings.Builder
	for _, character := range value {
		if character != '_' && character != '-' {
			builder.WriteRune(unicode.ToLower(character))
		}
	}
	return builder.String()
}

func credentialKey(key string) bool {
	_, ok := credentialKeys[key]
	return ok
}

func referenceCredentialKey(key string) bool {
	_, ok := referenceCredentialKeys[key]
	return ok
}

func exactStringSet(values ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func structuralCredentialReference(raw string) bool {
	value := unquoteScalar(raw)
	if environmentReferencePattern.MatchString(value) || structuredReferencePattern.MatchString(value) {
		return true
	}
	switch strings.ToLower(value) {
	case "<redacted>", "***redacted***", "<placeholder>":
		return true
	default:
		return false
	}
}

func validReferenceCredentialValue(key string, raw string) bool {
	value := unquoteScalar(raw)
	if environmentReferencePattern.MatchString(value) {
		return true
	}
	switch key {
	case "token" + "env":
		return environmentNamePattern.MatchString(value)
	case "secret" + "name":
		return dnsReferencePattern.MatchString(value)
	case "secret" + "ref", "secret" + "keyref", "secret" + "storeref", "credential" + "ref", "password" + "ref", "apikey" + "ref", "accesskey" + "ref", "clientsecret" + "ref", "privatekey" + "ref":
		return structuredReferencePattern.MatchString(value)
	default:
		return false
	}
}

func unquoteScalar(raw string) string {
	value := strings.TrimSpace(raw)
	if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
		return value[1 : len(value)-1]
	}
	return value
}

func privateEndpoint(line string) bool {
	if privateHostnamePattern.FindStringIndex(line) != nil {
		return true
	}
	lower := strings.ToLower(line)
	endpointContext := strings.Contains(lower, "endpoint") || strings.Contains(lower, "url") || strings.Contains(lower, "host:") || strings.Contains(lower, "host=")
	if endpointContext && barePrivateHostnamePattern.FindStringIndex(line) != nil {
		return true
	}
	for offset := 0; offset < len(line); {
		match := ipv4TokenPattern.FindStringIndex(line[offset:])
		if match == nil {
			break
		}
		addressText := line[offset+match[0] : offset+match[1]]
		offset += match[1]
		address, err := netip.ParseAddr(addressText)
		if err != nil || !address.Is4() {
			continue
		}
		if address.IsPrivate() || address.IsLoopback() || address.IsLinkLocalUnicast() || address.IsUnspecified() {
			return true
		}
	}
	for offset := 0; offset < len(line); {
		match := ipv6TokenPattern.FindStringIndex(line[offset:])
		if match == nil {
			break
		}
		addressText := line[offset+match[0] : offset+match[1]]
		offset += match[1]
		candidate := strings.Trim(addressText, "[]")
		if candidate == "::" {
			if endpointContext {
				return true
			}
			continue
		}
		address, err := netip.ParseAddr(candidate)
		if err != nil || !address.Is6() {
			continue
		}
		if address.IsPrivate() || address.IsLoopback() || address.IsLinkLocalUnicast() || address.IsUnspecified() {
			return true
		}
	}
	return false
}

func localFilesystemPath(line string) bool {
	return unixUserPathPattern.FindStringIndex(line) != nil || windowsUserPathPattern.FindStringIndex(line) != nil
}

func privateTreeReference(line string) bool {
	lower := strings.ToLower(line)
	privateModule := "cloudring" + ".local/" + "platform"
	privateEvidence := "." + "omo/"
	privateRequirements := "require" + "ments/"
	return strings.Contains(lower, privateModule) || strings.Contains(lower, privateEvidence) || strings.Contains(lower, privateRequirements)
}

func privateSourceAttribution(line string) bool {
	lower := strings.ToLower(strings.Join(strings.Fields(line), " "))
	phrases := []string{
		"copied " + "from private " + "source",
		"copied " + "from proprietary " + "source",
		"copied " + "from internal " + "repository",
		"ported " + "from private " + "repository",
	}
	for _, phrase := range phrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}
