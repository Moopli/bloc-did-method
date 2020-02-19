/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"errors"
)

// ChangeSet: a set of changes which are all applied together to a configuration
// A ChangeSet contains at least _some_ change to either consortium or stakeholder configs
// A ChangeSet always changes the consortium config, since this contains hashlinks to stakeholder configs.
// A ChangeSet contains the updated version of each changed file
type ChangeSet struct {
	// invariants:
	//  - consortium.Data is the marshalled json of consortium.Config
	//  - consortium.key is the hash of consortium.Data
	consortium *ConsortiumData
	// stakeholders maps from the hash key of a stakeholder config to the data object
	stakeholders       map[string]*StakeholderData
	stakeholderHistory map[string]*StakeholderData
	consortiumHistory  map[string]*ConsortiumData
}

// ApplyStakeholderChange Apply a stakeholder change to this ChangeSet
func (cs *ChangeSet) ApplyStakeholderChange(change *StakeholderData) error {
	if change.Config.Previous != "" {
		c, ok := cs.stakeholders[change.Config.Previous]
		// move the element to history
		// todo: figure out how to verify that a config can only be replaced by one which is endorsed by the same stakeholder
		//   (ie, a stakeholder can't replace someone else's)
		//   (this probably requires the endorsement/validation code to do the verification check)
		if ok && c != nil {
			cs.stakeholderHistory[change.Config.Previous] = c
			delete(cs.stakeholders, change.Config.Previous)
		} else {
			return errors.New("changeset missing predecessor to stakeholder change")
		}
	}

	cs.stakeholders[change.key] = change
	return nil
}

// ApplyConsortiumChange apply a consortium change to this ChangeSet
func (cs *ChangeSet) ApplyConsortiumChange(change *ConsortiumData) error {
	if change.Config.Previous != cs.consortium.key {
		return errors.New("changeset missing predecessor to consortium change")
	}

	cs.consortiumHistory[cs.consortium.key] = cs.consortium
	cs.consortium = change
	return nil
}

// RefreshConsortiumStakeholderLinks refreshes the stakeholder links in the consortium data
// TODO: change func name
func (cs *ChangeSet) RefreshConsortiumStakeholderLinks() error {
	newConf := cs.consortium.Config.Copy()
	newConf.Previous = cs.consortium.key

	newConf.Stakeholders = nil

	for _, sd := range cs.stakeholders {
		newConf.Stakeholders = append(newConf.Stakeholders, StakeholderListElement{
			Domain: sd.Config.Domain,
			Config: sd.key,
		})
	}

	consData, err := WrapConsortiumConf(newConf)
	if err != nil {
		return err
	}

	cs.consortiumHistory[cs.consortium.key] = cs.consortium
	cs.consortium = consData

	return nil
}

// NOTE: ConsortiumData needs to always be constructed with WrapConsortiumConf
//       to ensure that the key and Data always accurately reflect the config contents
type ConsortiumData struct {
	key    string
	Config *ConsortiumConfig
	Data   []byte
}

func WrapConsortiumConf(cc *ConsortiumConfig) (*ConsortiumData, error) {
	key, data, err := MarshalAndHash(cc)
	if err != nil {
		return nil, err
	}

	cd := &ConsortiumData{
		key:    key,
		Config: cc,
		Data:   data,
	}

	return cd, nil
}

type StakeholderData struct {
	key    string
	Config *StakeholderConfig
	Data   []byte
}

// maybe instead of merging we should have smaller change operations which are then applied onto a ChangeSet
// and then it's finalized and saved?

func WrapStakeholderConf(sc *StakeholderConfig) (*StakeholderData, error) {
	key, data, err := MarshalAndHash(sc)
	if err != nil {
		return nil, err
	}

	sd := &StakeholderData{
		key:    key,
		Config: sc,
		Data:   data,
	}

	return sd, nil
}

// EditStakeholder edits a StakeholderConfig, creating a new StakeholderConfig as a descendant
// This function takes an edit function and wraps it to edit the StakeholderData history object
func EditStakeholder(data StakeholderData, edFunc func(config *StakeholderConfig)) (*StakeholderData, error) {
	newConf := data.Config.Copy()
	edFunc(newConf)

	newConf.Previous = data.key

	newData, err := WrapStakeholderConf(newConf)
	if err != nil {
		return nil, err
	}

	return newData, nil
}
