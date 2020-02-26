/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

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

// StakeholderSettings holds the stakeholder settings
type StakeholderSettings struct {
	Cache CacheControl `json:"cache"`
}

// CacheControl holds cache settings for this file, indicating to the recipient how long until they should check for a new version of the file.
type CacheControl struct {
	MaxAge uint32 `json:"max-age"`
}
