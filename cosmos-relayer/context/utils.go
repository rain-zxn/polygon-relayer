/*
 * Copyright (C) 2020 The poly network Authors
 * This file is part of The poly network library.
 *
 * The  poly network  is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The  poly network  is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 * You should have received a copy of the GNU Lesser General Public License
 * along with The poly network .  If not, see <http://www.gnu.org/licenses/>.
 */

package context

import (
	"fmt"

	poly_go_sdk "github.com/polynetwork/poly-go-sdk"
	sdkp "github.com/polynetwork/polygon-relayer/poly_go_sdk"
)

func GetAccountByPassword(sdk *sdkp.PolySdk, path string, pwd []byte) (*poly_go_sdk.Account, error) {
	wallet, err := sdk.OpenWallet(path)
	if err != nil {
		return nil, fmt.Errorf("open wallet error: %v", err)
	}
	user, err := wallet.GetDefaultAccount(pwd)
	if err != nil {
		return nil, fmt.Errorf("getDefaultAccount error: %v", err)
	}
	return user, nil
}
