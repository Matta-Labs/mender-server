// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package main

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/log"

	api_http "github.com/mendersoftware/mender-server/services/useradm/api/http"
	"github.com/mendersoftware/mender-server/services/useradm/client/tenant"
	"github.com/mendersoftware/mender-server/services/useradm/common"
	. "github.com/mendersoftware/mender-server/services/useradm/config"
	"github.com/mendersoftware/mender-server/services/useradm/jwt"
	"github.com/mendersoftware/mender-server/services/useradm/store/mongo"
	useradm "github.com/mendersoftware/mender-server/services/useradm/user"
)

func RunServer(c config.Reader) error {

	l := log.New(log.Ctx{})

	authorizer := &SimpleAuthz{}

	// let's now go through all the existing keys and load them
	jwtHandlers, err := addPrivateKeys(
		l,
		filepath.Dir(c.GetString(SettingServerPrivKeyPath)),
		c.GetString(SettingServerPrivKeyFileNamePattern),
	)
	if err != nil {
		return err
	}

	// the handler for keyId equal 0 is the one associated with the
	// SettingServerPrivKeyPathDefault key. it is the one serving all the previously
	// issued tokens (before the kid introduction in the JWTs)
	defaultHandler, err := jwt.NewJWTHandler(
		SettingServerPrivKeyPathDefault,
		c.GetString(SettingServerPrivKeyFileNamePattern),
	)
	if err == nil && defaultHandler != nil {
		// the key with id 0 is by default the default one. this allows
		// to support tokens without "kid" in the header
		// it is possible, that you rotated the default key, in which case you have to
		// set USERADM_SERVER_PRIV_KEY_PATH=/etc/useradm/rsa/private.id.2048.pem
		// where private.id.2048.pem is the new key, with new id. the new one will by default
		// be used to issue new tokens, while any other token which has id that we have
		// will be authorized against its matching key (by id from "kid" in JWT header)
		// or which does not have "kid" will be authorized against the key with id 0.
		// in other words: the key with id 0 (if not present as private.id.0.pem)
		// is the default one, and all the JWT with no "kid" in headers are being
		// checked against it.
		jwtHandlers[common.KeyIdZero] = defaultHandler
	}

	// if the default path is different from the currently set key path
	// we still have not loaded this key. this happens when the key rotation took place,
	// someone exported USERADM_SERVER_PRIV_KEY_PATH=path-to-a-new-key and this key
	// now will serve all. if we do not have this, the Login will fall back to the keyId
	// from the filename and either use the KeyIdZero key or fail to find the key to issue a
	// token if the one set in USERADM_SERVER_PRIV_KEY_PATH does have id in the filename
	// (but does not exist because we have not loaded it)
	// this also means that careless setting of USERADM_SERVER_PRIV_KEY_PATH to a key that does
	// not match the SettingServerPrivKeyFileNamePattern will result in
	// KeyIdZero handler overwrite and lack of back support for tokens signed by it.
	if c.GetString(SettingServerPrivKeyPath) != SettingServerPrivKeyPathDefault {
		defaultHandler, err = jwt.NewJWTHandler(
			c.GetString(SettingServerPrivKeyPath),
			c.GetString(SettingServerPrivKeyFileNamePattern),
		)
		if err == nil && defaultHandler != nil {
			keyId := common.KeyIdFromPath(
				c.GetString(SettingServerPrivKeyPath),
				c.GetString(SettingServerPrivKeyFileNamePattern),
			)
			if keyId == common.KeyIdZero {
				l.Warnf(
					"currently set private key %s either does not match %s pattern"+
						" or has explicitly set id=0. we are overridding the default"+
						" private key handler with id=0",
					c.GetString(SettingServerPrivKeyPath),
					c.GetString(SettingServerPrivKeyFileNamePattern),
				)
			}
			jwtHandlers[keyId] = defaultHandler
		}
	}

	var jwtFallbackHandler jwt.Handler
	fallback := c.GetString(SettingServerFallbackPrivKeyPath)
	if err == nil && fallback != "" {
		jwtFallbackHandler, err = jwt.NewJWTHandler(
			fallback,
			c.GetString(SettingServerPrivKeyFileNamePattern),
		)
	}
	if err != nil {
		return err
	}

	db, err := mongo.GetDataStoreMongo(dataStoreMongoConfigFromAppConfig(c))
	if err != nil {
		return errors.Wrap(err, "database connection failed")
	}

	ua := useradm.NewUserAdm(jwtHandlers, db,
		useradm.Config{
			Issuer:                         c.GetString(SettingJWTIssuer),
			ExpirationTimeSeconds:          int64(c.GetInt(SettingJWTExpirationTimeout)),
			LimitSessionsPerUser:           c.GetInt(SettingLimitSessionsPerUser),
			LimitTokensPerUser:             c.GetInt(SettingLimitTokensPerUser),
			TokenLastUsedUpdateFreqMinutes: c.GetInt(SettingTokenLastUsedUpdateFreqMinutes),
			PrivateKeyPath:                 c.GetString(SettingServerPrivKeyPath),
			PrivateKeyFileNamePattern:      c.GetString(SettingServerPrivKeyFileNamePattern),
		})

	if tadmAddr := c.GetString(SettingTenantAdmAddr); tadmAddr != "" {
		l.Infof("settting up tenant verification")

		tc := tenant.NewClient(tenant.Config{
			TenantAdmAddr: tadmAddr,
		})

		ua = ua.WithTenantVerification(tc)
	}

	useradmapi := api_http.NewUserAdmApiHandlers(ua, db, jwtHandlers,
		api_http.Config{
			TokenMaxExpSeconds: c.GetInt(SettingTokenMaxExpirationSeconds),
			JWTFallback:        jwtFallbackHandler,
		})

	handler, err := useradmapi.Build(authorizer)
	if err != nil {
		return errors.Wrap(err, "useradm API handlers setup failed")
	}

	addr := c.GetString(SettingListen)
	l.Printf("listening on %s", addr)

	return http.ListenAndServe(addr, handler)
}

func addPrivateKeys(
	l *log.Logger,
	privateKeysDirectory string,
	privateKeyPattern string,
) (handlers map[int]jwt.Handler, err error) {
	files, err := os.ReadDir(privateKeysDirectory)
	if err != nil {
		return
	}

	r, err := regexp.Compile(privateKeyPattern)
	if err != nil {
		return
	}

	handlers = make(map[int]jwt.Handler, len(files))
	for _, fileEntry := range files {
		if r.MatchString(fileEntry.Name()) {
			keyPath := path.Join(privateKeysDirectory, fileEntry.Name())
			handler, err := jwt.NewJWTHandler(keyPath, privateKeyPattern)
			if err != nil {
				continue
			}
			keyId := common.KeyIdFromPath(keyPath, privateKeyPattern)
			l.Infof("loaded private key id=%d from %s", keyId, keyPath)
			handlers[keyId] = handler
		}
	}
	return handlers, nil
}
