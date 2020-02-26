/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChangeSet(t *testing.T) {

	t.Run("success: construct a changeset, apply multiple changes to a stakeholder", func(t *testing.T) {
		sc, err := WrapStakeholderConf(&StakeholderConfig{
			Domain:    "stakeholder.1.v1",
			Previous:  "<previous stakeholder>",
			Endpoints: []string{},
		})
		require.NoError(t, err)

		cc, err := WrapConsortiumConf(&ConsortiumConfig{
			Domain:   "consortium",
			Previous: "<previous consortium>",
		})
		require.NoError(t, err)

		cs := ChangeSet{
			consortium:         cc,
			stakeholders:       map[string]*StakeholderData{},
			stakeholderHistory: map[string]*StakeholderData{},
			consortiumHistory:  map[string]*ConsortiumData{},
		}
		cs.stakeholders[sc.key] = sc

		err = cs.RefreshConsortiumStakeholderLinks()
		require.NoError(t, err)

		sc2, err := EditStakeholder(sc, func(config *StakeholderConfig) {
			config.Domain = "stakeholder.1.v2"
		})
		require.NoError(t, err)

		err = cs.ApplyStakeholderChange(sc2)
		require.NoError(t, err)

		err = cs.RefreshConsortiumStakeholderLinks()
		require.NoError(t, err)

		sc3, err := EditStakeholder(sc2, func(config *StakeholderConfig) {
			config.Domain = "stakeholder.1.v3"
		})
		require.NoError(t, err)

		err = cs.ApplyStakeholderChange(sc3)
		require.NoError(t, err)

		err = cs.RefreshConsortiumStakeholderLinks()
		require.NoError(t, err)

		println("--------- before squash: ---------")
		println("consortium:")
		println(cs.consortium.key, ":", cs.consortium.Config.Domain, ":", cs.consortium.Config.Previous)
		println("consortium history:")
		for k, v := range cs.consortiumHistory {
			println(k, ":", v.Config.Domain, ":", v.Config.Previous)
		}
		println("stakeholders:")
		for k, v := range cs.stakeholders {
			println(k, ":", v.Config.Domain, ":", v.Config.Previous)
		}
		println("stakeholder history:")
		for k, v := range cs.stakeholderHistory {
			println(k, ":", v.Config.Domain, ":", v.Config.Previous)
		}

		err = cs.SquashHistory()
		require.NoError(t, err)

		println("---------  after squash: ---------")
		println("consortium:")
		println(cs.consortium.key, ":", cs.consortium.Config.Domain, ":", cs.consortium.Config.Previous)
		println("consortium history:")
		for k, v := range cs.consortiumHistory {
			println(k, ":", v.Config.Domain, ":", v.Config.Previous)
		}
		println("stakeholders:")
		for k, v := range cs.stakeholders {
			println(k, ":", v.Config.Domain, ":", v.Config.Previous)
		}
		println("stakeholder history:")
		for k, v := range cs.stakeholderHistory {
			println(k, ":", v.Config.Domain, ":", v.Config.Previous)
		}
	})

	t.Run("compare EditStakeholder against direct editing", func(t *testing.T) {
		sc, err := WrapStakeholderConf(&StakeholderConfig{
			Domain:    "stakeholder.1.v1",
			Previous:  "<previous stakeholder>",
			Endpoints: []string{},
		})
		require.NoError(t, err)

		sc2test, err := EditStakeholder(sc, func(config *StakeholderConfig) {
			config.Domain = "stakeholder.1.v2"
		})
		require.NoError(t, err)

		println(sc2test.key, ":", string(sc2test.Data))

		sc2, err := WrapStakeholderConf(&StakeholderConfig{
			Domain:    "stakeholder.1.v2",
			Previous:  sc.key,
			Endpoints: []string{},
		})
		require.NoError(t, err)

		println(sc2.key, ":", string(sc2.Data))

		require.Equal(t, sc2.key, sc2test.key)
	})

}
