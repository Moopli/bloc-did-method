/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"errors"
)

/*
ChangeSet is used to construct a set of changes that will be applied all together to a configuration.

Adding a consortium config (to a ChangeSet without one)
- Construct a new ConsortiumConfig, set its values as necessary
- changeSet.NewConsortium(the new config)
Editing the consortium config:
- changeSet.EditConsortium(<editing callback>)
  - you provide a callback that edits a copy of the consortium config, which will become the new consortium config (while the original is saved in history)

Adding a stakeholder config:
- Construct a new StakeholderConfig, set its values as necessary
- changeSet.NewStakeholder(the new config)
Editing a stakeholder config:
- changeSet.EditStakeholder(<key of config to replace>, <editing callback>) -> <key of new config>
  - you provide a callback that edits a copy of the chosen stakeholder config, which will replace that config

Bookkeeping:
- SquashHistory(): when a ChangeSet is ready to be signed or applied, squash the history in the changeset to avoid clutter
- RefreshConsortiumStakeholderLinks(): after stakeholders are changed, update the consortium config so it points to the current stakeholder configs instead of historical ones

TODO: DeleteStakeholder() function
TODO: Signing/Endorsement of ChangeSets

TODO: un-expose any helper functions that don't need exposing
*/

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

// NewConsortium creates a new consortium config inside this ChangeSet, as long as one does not already exist
// The parameter cc is added to the ChangeSet as the new config
func (cs *ChangeSet) NewConsortium(cc *ConsortiumConfig) error {
	if cs.consortium != nil {
		return errors.New("changeset already contains consortium")
	}

	cd, err := WrapConsortiumConf(cc)
	if err != nil {
		return err
	}

	return cs.ApplyConsortiumChange(cd)
}

// EditConsortium edits the current consortium configuration of the ChangeSet
func (cs *ChangeSet) EditConsortium(edFunc func(config *ConsortiumConfig)) error {
	newConf := cs.consortium.Config.Copy()
	edFunc(newConf)

	newConf.Previous = cs.consortium.key

	cd, err := WrapConsortiumConf(newConf)
	if err != nil {
		return err
	}

	return cs.ApplyConsortiumChange(cd)
}

// NewStakeholder creates a new stakeholder config inside this ChangeSet, as long as it isn't already in the ChangeSet or its history
func (cs *ChangeSet) NewStakeholder(sc *StakeholderConfig) error {
	sd, err := WrapStakeholderConf(sc)
	if err != nil {
		return err
	}

	if _, ok := cs.stakeholders[sd.key]; ok {
		return errors.New("can't add duplicate stakeholder config as new config")
	}
	if _, ok := cs.stakeholderHistory[sd.key]; ok {
		// given the existence of `prev`, even a change that reverts a config to a prior value is going to have a different ID
		return errors.New("can't add stakeholder config from history as new config")
	}

	return cs.ApplyStakeholderChange(sd)
}

// EditStakeholder edits one of the stakeholders of the ChangeSet, moving the original into the history
func (cs *ChangeSet) EditStakeholder(data *StakeholderData, edFunc func(config *StakeholderConfig)) (string, error) {
	newConf := data.Config.Copy()
	edFunc(newConf)

	newConf.Previous = data.key

	sd, err := WrapStakeholderConf(newConf)
	if err != nil {
		return "", err
	}

	err = cs.ApplyStakeholderChange(sd)
	if err != nil {
		return "", err
	}

	return sd.key, nil
}

var InconsistentChangeSet = errors.New("inconsistent changeset")

// SquashHistory squashes the history in the ChangeSet, so the final state derives directly from the initial history
func (cs *ChangeSet) SquashHistory() error {
	// traversedHistory is a set indicating whether a stakeholder history object has been traversed yet, to catch history branches
	traversedHistory := make(map[string]struct{})

	// Add all current stakeholder configs to traversedHistory, since none of these are allowed to be the predecessor of anything
	for key, _ := range cs.stakeholders {
		traversedHistory[key] = struct{}{}
	}

	// For each stakeholder config, traverse its ancestors, and for each ancestor:
	//  - check traversedHistory if this ancestor was already in there
	//  - log them in traversedHistory, so no subsequent stakeholder config can share the ancestor
	//  -
	for _, currentStakeholder := range cs.stakeholders {
		next := currentStakeholder.Config.Previous
		var ok = true

		for ok {
			if _, present := traversedHistory[next]; present {
				// `next` was already present in the traversed history - so another stakeholder reached it first. Duplicate.
				return InconsistentChangeSet
			}

			var val *StakeholderData
			// val.key == next
			val, ok = cs.stakeholderHistory[next]

			if ok {
				traversedHistory[next] = struct{}{}
				// remove the history element that has been bypassed (from stakeholderHistory)
				delete(cs.stakeholderHistory, next)
				next = val.Config.Previous
			} else {
				// Reached the end of the history chain in this ChangeSet
				// so we repoint the current stakeholder's ancestor to be the first ancestor outside this changeset
				currentStakeholder.Config.Previous = next
				// and editing its value means we must regenerate its json serialization and hash
				hash, data, err := MarshalAndHash(currentStakeholder.Config)
				if err != nil {
					return err
				}
				currentStakeholder.key = hash
				currentStakeholder.Data = data
			}
		}
	}

	// Having edited stakeholders, we update the consortium to contain correct links to the stakeholders
	err := cs.RefreshConsortiumStakeholderLinks()
	if err != nil {
		return err
	}

	// Squash consortium history
	next := cs.consortium.Config.Previous
	var ok = true

	for ok {
		var val *ConsortiumData
		// val.key == next
		val, ok = cs.consortiumHistory[next]

		if ok {
			// remove the history element that has been bypassed (from consortiumHistory)
			delete(cs.consortiumHistory, next)
			next = val.Config.Previous
		} else {
			// Reached the end of the consortium history chain in this ChangeSet
			cs.consortium.Config.Previous = next
			// and editing its value means we must regenerate its json serialization and hash
			hash, data, err := MarshalAndHash(cs.consortium.Config)
			if err != nil {
				return err
			}
			cs.consortium.key = hash
			cs.consortium.Data = data
		}
	}

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

	return cs.ApplyConsortiumChange(consData)
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

	return &ConsortiumData{
		key:    key,
		Config: cc,
		Data:   data,
	}, nil
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

	return &StakeholderData{
		key:    key,
		Config: sc,
		Data:   data,
	}, nil
}

// EditStakeholder edits a StakeholderConfig, creating a new StakeholderConfig as a descendant
func EditStakeholder(data *StakeholderData, edFunc func(config *StakeholderConfig)) (*StakeholderData, error) {
	newConf := data.Config.Copy()
	edFunc(newConf)

	newConf.Previous = data.key

	return WrapStakeholderConf(newConf)
}
