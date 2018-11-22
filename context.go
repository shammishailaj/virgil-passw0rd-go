/*
 * Copyright (C) 2015-2018 Virgil Security Inc.
 *
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 * met:
 *
 *     (1) Redistributions of source code must retain the above copyright
 *     notice, this list of conditions and the following disclaimer.
 *
 *     (2) Redistributions in binary form must reproduce the above copyright
 *     notice, this list of conditions and the following disclaimer in
 *     the documentation and/or other materials provided with the
 *     distribution.
 *
 *     (3) Neither the name of the copyright holder nor the names of its
 *     contributors may be used to endorse or promote products derived from
 *     this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE AUTHOR ''AS IS'' AND ANY EXPRESS OR
 * IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY DIRECT,
 * INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 * SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
 * HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
 * STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING
 * IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 *
 * Lead Maintainer: Virgil Security Inc. <support@virgilsecurity.com>
 */

package passw0rd

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/passw0rd/phe-go"

	"github.com/pkg/errors"
)

type Context struct {
	AccessToken  string
	AppId        string
	PHEClients   map[uint32]*phe.Client
	UpdateTokens map[uint32][]byte
	Version      uint32
}

func CreateContext(accessToken, appId, clientSecretKey, serverPublicKey string, updateTokens ...string) (*Context, error) {

	if len(appId) != 32 || clientSecretKey == "" || serverPublicKey == "" || accessToken == "" {
		return nil, errors.New("all parameters are mandatory")
	}

	_, err := hex.DecodeString(appId)
	if err != nil {
		return nil, errors.New("invalid appID")
	}

	skVersion, sk, err := ParseVersionAndContent("SK", clientSecretKey)
	if err != nil {
		return nil, errors.Wrap(err, "invalid secret key")
	}

	pubVersion, pubBytes, err := ParseVersionAndContent("PK", serverPublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "invalid public key")
	}

	if skVersion != pubVersion {
		return nil, errors.New("public and secret keys must have the same version")
	}

	currentSk, currentPub := sk, pubBytes
	pheClient, err := phe.NewClient(currentSk, currentPub)

	if err != nil {
		return nil, errors.Wrap(err, "could not create PHE client")
	}

	phes := make(map[uint32]*phe.Client)
	phes[pubVersion] = pheClient

	tokens, err := parseTokens(updateTokens...)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse update tokens")
	}

	currentVersion := pubVersion

	var tokenMap map[uint32][]byte

	if len(tokens) > 0 {
		tokenMap = make(map[uint32][]byte)
		for _, token := range tokens {
			if token.Version != currentVersion+1 {
				return nil, fmt.Errorf("incorrect token version %d", token.Version)
			}

			nextSk, nextPub, err := phe.RotateClientKeys(currentSk, currentPub, token.UpdateToken)
			if err != nil {
				return nil, errors.Wrap(err, "could not update keys using token")
			}

			nextClient, err := phe.NewClient(nextSk, nextPub)
			if err != nil {
				return nil, errors.Wrap(err, "could not create PHE client")
			}

			phes[token.Version] = nextClient
			currentSk, currentPub = nextSk, nextPub
			currentVersion = token.Version
			tokenMap[token.Version] = token.UpdateToken
		}

	}

	return &Context{
		AccessToken:  accessToken,
		PHEClients:   phes,
		AppId:        appId,
		Version:      currentVersion,
		UpdateTokens: tokenMap,
	}, nil
}

func parseTokens(tokens ...string) (parsedTokens []*VersionedUpdateToken, err error) {
	if len(tokens) == 0 {
		return nil, nil
	}

	for _, tokenStr := range tokens {

		version, content, err := ParseVersionAndContent("UT", tokenStr)

		if err != nil {
			return nil, errors.Wrap(err, "invalid update token")
		}

		vt := &VersionedUpdateToken{
			Version:     version,
			UpdateToken: content,
		}

		parsedTokens = append(parsedTokens, vt)
	}

	sort.Slice(parsedTokens, func(i, j int) bool { return parsedTokens[i].Version < parsedTokens[j].Version })

	return
}

func ParseVersionAndContent(prefix, str string) (version uint32, content []byte, err error) {
	parts := strings.Split(str, ".")
	if len(parts) != 3 || parts[0] != prefix {
		return 0, nil, errors.New("invalid string")
	}

	nVersion, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, nil, errors.Wrap(err, "invalid string")
	}

	if version < 1 {
		return 0, nil, errors.Wrap(err, "invalid version")
	}
	version = uint32(nVersion)

	content, err = base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return 0, nil, errors.Wrap(err, "invalid string")
	}
	return
}
