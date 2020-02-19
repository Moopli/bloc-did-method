/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"encoding/json"

	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multihash"
)

// Models for config files

/*
The consortium config contains the following:
- The domain name of the consortium
- Consortium policy configuration settings
- A list of stakeholders - containing, for each stakeholder:
  - The web domain where its configuration can be found
  - A local hashlink to the cached stakeholder configuration
- A list of detached JWS signatures, from the stakeholders, on the config file
- A local hashlink to the previous version of this config file
*/

type ConsortiumConfig struct {
	// Domain is the domain name of the consortium
	Domain string `json:"domain,omitempty"`
	// Policy contains the consortium policy configuration
	Policy ConsortiumPolicy `json:"policy"`
	// Stakeholders is a list containing references to the stakeholders on this consortium
	Stakeholders []StakeholderListElement `json:"stakeholders"`
	// Signatures is a list of detached JWSs signed by the stakeholders on this configuration
	Signatures []DetachedJWS `json:"signatures"`
	// Previous contains a hashlink to the previous version of this file. Optional.
	Previous string `json:"previous,omitempty"`
}

func (cc *ConsortiumConfig) Copy() *ConsortiumConfig {
	stakeHolders := make([]StakeholderListElement, len(cc.Stakeholders))
	copy(stakeHolders, cc.Stakeholders)

	signatures := make([]DetachedJWS, len(cc.Signatures))
	copy(signatures, cc.Signatures)

	return &ConsortiumConfig{
		Domain:       cc.Domain,
		Policy:       cc.Policy,
		Stakeholders: stakeHolders,
		Signatures:   signatures,
		Previous:     cc.Previous,
	}
}

type ConsortiumPolicy struct {
	Cache CacheControl `json:"cache"`
}

type StakeholderListElement struct {
	// Domain is the domain name of the stakeholder
	Domain string `json:"domain,omitempty"`
	// Config is a hashlink to a local copy of the stakeholder configuration file
	Config string `json:"config,omitempty"`
}

type DetachedJWS struct {
}

/*
A stakeholder configuration file contains:
- The stakeholder's domain
- Stakeholder custom configuration settings
- The stakeholder's Sidetree endpoints
- A detached JWS signature on the file, created by the stakeholder
- a local hashlink to the previous version of this config file
*/

// StakeholderConfig holds the configuration for a stakeholder
type StakeholderConfig struct {
	// Domain is the domain name of the stakeholder organisation, where the primary copy of the stakeholder config can be found
	Domain string `json:"domain,omitempty"`
	// Config contains stakeholder-specific configuration settings
	Config StakeholderSettings `json:"conf"`
	// Endpoints is a list of sidetree endpoints owned by this stakeholder organization
	Endpoints []string    `json:"endpoints"`
	Signature DetachedJWS `json:"sig,omitempty"`
	// Previous is a hashlink to the previous version of this file
	Previous string `json:"previous,omitempty"`
}

func (sc *StakeholderConfig) Copy() *StakeholderConfig {
	endPoints := make([]string, len(sc.Endpoints))
	copy(endPoints, sc.Endpoints)

	return &StakeholderConfig{
		Domain:    sc.Domain,
		Config:    sc.Config.Copy(),
		Endpoints: endPoints,
		Signature: sc.Signature,
		Previous:  sc.Previous,
	}
}

// StakeholderSettings holds the stakeholder settings
type StakeholderSettings struct {
	Cache CacheControl `json:"cache"`
}

func (ss *StakeholderSettings) Copy() StakeholderSettings {
	return StakeholderSettings{
		Cache: ss.Cache,
	}
}

// CacheControl holds cache settings for this file, indicating to the recipient how long until they should check for a new version of the file.
type CacheControl struct {
	MaxAge uint32 `json:"max-age"`
}

// MarshalAndHash marshals a json struct and returns the marshaled bytes and the hash of the data
func MarshalAndHash(data interface{}) (string, []byte, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", nil, err
	}

	hash, err := multihash.Sum(bytes, multihash.SHA2_256, -1)
	if err != nil {
		return "", nil, err
	}

	key, err := multibase.Encode(multibase.Base58BTC, hash)
	if err != nil {
		return "", nil, err
	}

	return key, bytes, nil
}
