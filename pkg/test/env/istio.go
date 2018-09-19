//  Copyright 2018 Istio Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package env

import (
	"fmt"
	"go/build"
	"os"
	"strings"

	"runtime"

	"istio.io/istio/pkg/log"
)

var (
	// TOP environment variable
	// nolint: golint
	TOP Variable = "TOP"

	// ISTIO_GO environment variable
	// nolint: golint
	ISTIO_GO Variable = "ISTIO_GO"

	// ISTIO_BIN environment variable
	// nolint: golint
	ISTIO_BIN Variable = "ISTIO_BIN"

	// ISTIO_OUT environment variable
	// nolint: golint
	ISTIO_OUT Variable = "ISTIO_OUT"

	// IstioTop has the top of the istio tree, matches the env variable from make.
	IstioTop = TOP.ValueOrDefaultFunc(getDefaultIstioTop)

	// IstioSrc is the location if istio source ($TOP/src/istio.io/istio
	IstioSrc = IstioTop + "/src/istio.io/istio"

	// IstioBin is the location of the binary output directory
	IstioBin = verifyFile(ISTIO_BIN, ISTIO_BIN.ValueOrDefaultFunc(getDefaultIstioBin))

	// IstioOut is the location of the output directory ($TOP/out)
	IstioOut = verifyFile(ISTIO_OUT, ISTIO_OUT.ValueOrDefaultFunc(getDefaultIstioOut))
)

func getDefaultIstioTop() string {
	// Assume it is run inside istio.io/istio
	current, _ := os.Getwd()
	idx := strings.Index(current, "/src/istio.io/istio")
	if idx > 0 {
		return current[0:idx]
	}
	return current // launching from GOTOP (for example in goland)
}

func getDefaultIstioBin() string {
	return fmt.Sprintf("%s/bin", build.Default.GOPATH)
}

func getDefaultIstioOut() string {
	return fmt.Sprintf("%s/out/%s_%s", build.Default.GOPATH, runtime.GOOS, runtime.GOARCH)
}

func verifyFile(v Variable, f string) string {
	if !fileExists(f) {
		log.Warnf("unable to resolve %s. Dir %s does not exist", v, f)
		return ""
	}
	return f
}

func fileExists(f string) bool {
	_, err := os.Stat(f)
	return !os.IsNotExist(err)
}
