package avm

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	MaxInterfaceNameLength = 64
	MaxInterfaceMethods    = 128
	MaxInterfaceEvents     = 128
)

type InterfaceManifest struct {
	Name    string
	Version uint16
	Methods []InterfaceMethod
	Events  []InterfaceEvent
}

type InterfaceMethod struct {
	Name       string
	Entrypoint Entrypoint
	Opcode     uint32
	Async      bool
}

type InterfaceEvent struct {
	Name   string
	Opcode uint32
}

func (m InterfaceManifest) Validate() error {
	if err := validateInterfaceName("interface", m.Name); err != nil {
		return err
	}
	if m.Version == 0 {
		return errors.New("AVM interface version must be positive")
	}
	if len(m.Methods) == 0 {
		return errors.New("AVM interface must declare at least one method")
	}
	if len(m.Methods) > MaxInterfaceMethods {
		return fmt.Errorf("AVM interface methods must be <= %d", MaxInterfaceMethods)
	}
	if len(m.Events) > MaxInterfaceEvents {
		return fmt.Errorf("AVM interface events must be <= %d", MaxInterfaceEvents)
	}
	seenMethods := make(map[string]struct{}, len(m.Methods))
	seenOpcodes := make(map[uint32]struct{}, len(m.Methods))
	for _, method := range m.Methods {
		if err := validateInterfaceName("interface method", method.Name); err != nil {
			return err
		}
		if !IsValidEntrypoint(method.Entrypoint) {
			return fmt.Errorf("AVM interface method %q has invalid entrypoint", method.Name)
		}
		if _, exists := seenMethods[method.Name]; exists {
			return fmt.Errorf("duplicate AVM interface method %q", method.Name)
		}
		seenMethods[method.Name] = struct{}{}
		if method.Opcode != 0 {
			if _, exists := seenOpcodes[method.Opcode]; exists {
				return fmt.Errorf("duplicate AVM interface opcode %d", method.Opcode)
			}
			seenOpcodes[method.Opcode] = struct{}{}
		}
	}
	seenEvents := make(map[string]struct{}, len(m.Events))
	for _, event := range m.Events {
		if err := validateInterfaceName("interface event", event.Name); err != nil {
			return err
		}
		if _, exists := seenEvents[event.Name]; exists {
			return fmt.Errorf("duplicate AVM interface event %q", event.Name)
		}
		seenEvents[event.Name] = struct{}{}
	}
	return nil
}

func InterfaceHash(manifest InterfaceManifest) ([MetadataHashLength]byte, error) {
	if err := manifest.Validate(); err != nil {
		return [MetadataHashLength]byte{}, err
	}
	manifest = canonicalInterfaceManifest(manifest)
	buf := bytes.NewBuffer(nil)
	writeString(buf, manifest.Name)
	writeU16(buf, manifest.Version)
	writeU16(buf, uint16(len(manifest.Methods)))
	for _, method := range manifest.Methods {
		writeString(buf, method.Name)
		buf.WriteByte(byte(method.Entrypoint))
		writeU32(buf, method.Opcode)
		if method.Async {
			buf.WriteByte(1)
		} else {
			buf.WriteByte(0)
		}
	}
	writeU16(buf, uint16(len(manifest.Events)))
	for _, event := range manifest.Events {
		writeString(buf, event.Name)
		writeU32(buf, event.Opcode)
	}
	return sha256.Sum256(buf.Bytes()), nil
}

func VerifyInterface(module Module, manifest InterfaceManifest) error {
	hash, err := InterfaceHash(manifest)
	if err != nil {
		return err
	}
	if !bytes.Equal(module.MetadataHash[:], hash[:]) {
		return errors.New("AVM interface metadata hash mismatch")
	}
	for _, method := range manifest.Methods {
		if _, ok := module.Exports[method.Entrypoint]; !ok {
			return fmt.Errorf("AVM interface method %q entrypoint is not exported", method.Name)
		}
	}
	return nil
}

func canonicalInterfaceManifest(manifest InterfaceManifest) InterfaceManifest {
	manifest.Name = strings.TrimSpace(manifest.Name)
	manifest.Methods = append([]InterfaceMethod(nil), manifest.Methods...)
	manifest.Events = append([]InterfaceEvent(nil), manifest.Events...)
	sort.SliceStable(manifest.Methods, func(i, j int) bool {
		if manifest.Methods[i].Name != manifest.Methods[j].Name {
			return manifest.Methods[i].Name < manifest.Methods[j].Name
		}
		if manifest.Methods[i].Entrypoint != manifest.Methods[j].Entrypoint {
			return manifest.Methods[i].Entrypoint < manifest.Methods[j].Entrypoint
		}
		return manifest.Methods[i].Opcode < manifest.Methods[j].Opcode
	})
	sort.SliceStable(manifest.Events, func(i, j int) bool {
		if manifest.Events[i].Name != manifest.Events[j].Name {
			return manifest.Events[i].Name < manifest.Events[j].Name
		}
		return manifest.Events[i].Opcode < manifest.Events[j].Opcode
	})
	return manifest
}

func validateInterfaceName(kind, name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("AVM %s name is required", kind)
	}
	if len(trimmed) > MaxInterfaceNameLength {
		return fmt.Errorf("AVM %s name must be <= %d bytes", kind, MaxInterfaceNameLength)
	}
	return nil
}
