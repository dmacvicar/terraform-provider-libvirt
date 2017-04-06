// +build api

/*
 * This file is part of the libvirt-go project
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 *
 * Copyright (c) 2013 Alex Zorin
 * Copyright (C) 2016 Red Hat, Inc.
 *
 */

package libvirt

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"testing"
)

/*
 * We search for 'C.XXXXX' in *.go. Any false positives need
 * to be listed in one of these vars below
 */
var (
	ignoreFuncs = []string{
		/* Obsolete we use virConnectDomainEventRegisterAny instead */
		"virConnectDomainEventRegister",

		/* Wrapped in connect_cfuncs.go instead */
		"virConnectOpenAuth",
		"virConnectRegisterCloseCallback",

		/* Auth callback typedef */
		"virConnectAuthCallbackPtr",

		/* Not thread safe, so not exposed */
		"virConnCopyLastError",
		"virConnGetLastError",
		"virConnResetLastError",
		"virConnSetErrorFunc",

		/* Only needed at C level */
		"virCopyLastError",
		"virFreeError",
		"virGetLastErrorMessage",
		"virResetLastError",
		"virSaveLastError",
		"virDefaultErrorFunc",

		/* Deprecated in favour of virDomainCreateXML */
		"virDomainCreateLinux",

		/* Connect callback typedef */
		"virConnectCloseFunc",

		/* Data free callback typedef */
		"virFreeCallback",

		/* Error callback typedef */
		"virErrorFunc",

		/* Domain event callback typedefs */
		"virConnectDomainEventAgentLifecycleCallback",
		"virConnectDomainEventBalloonChangeCallback",
		"virConnectDomainEventBlockJobCallback",
		"virConnectDomainEventCallback",
		"virConnectDomainEventDeregister",
		"virConnectDomainEventDeviceAddedCallback",
		"virConnectDomainEventDeviceRemovalFailedCallback",
		"virConnectDomainEventDeviceRemovedCallback",
		"virConnectDomainEventDiskChangeCallback",
		"virConnectDomainEventGraphicsCallback",
		"virConnectDomainEventIOErrorCallback",
		"virConnectDomainEventIOErrorReasonCallback",
		"virConnectDomainEventJobCompletedCallback",
		"virConnectDomainEventMigrationIterationCallback",
		"virConnectDomainEventPMSuspendCallback",
		"virConnectDomainEventPMSuspendDiskCallback",
		"virConnectDomainEventPMWakeupCallback",
		"virConnectDomainEventRTCChangeCallback",
		"virConnectDomainEventTrayChangeCallback",
		"virConnectDomainEventTunableCallback",
		"virConnectDomainEventWatchdogCallback",
		"virConnectDomainEventMetadataChangeCallback",
		"virConnectDomainEventBlockThresholdCallback",
		"virConnectDomainQemuMonitorEventCallback",

		/* Network event callback typedefs */
		"virConnectNetworkEventGenericCallback",
		"virConnectNetworkEventLifecycleCallback",

		/* Node device event callback typedefs */
		"virConnectNodeDeviceEventGenericCallback",
		"virConnectNodeDeviceEventLifecycleCallback",

		/* Secret event callback typedefs */
		"virConnectSecretEventGenericCallback",
		"virConnectSecretEventLifecycleCallback",

		/* Storage pool event callback typedefs */
		"virConnectStoragePoolEventGenericCallback",
		"virConnectStoragePoolEventLifecycleCallback",

		/* Stream event callback typedef */
		"virStreamEventCallback",

		/* Event loop callback typedefs */
		"virEventAddHandleFunc",
		"virEventAddTimeoutFunc",
		"virEventHandleCallback",
		"virEventRemoveHandleFunc",
		"virEventRemoveTimeoutFunc",
		"virEventTimeoutCallback",
		"virEventUpdateHandleFunc",
		"virEventUpdateTimeoutFunc",

		/* Typedefs that don't need exposing as is */
		"virStreamSinkFunc",
		"virStreamSourceFunc",
		"virStreamSinkHoleFunc",
		"virStreamSourceHoleFunc",
		"virStreamSourceSkipFunc",

		/* Only needed at C level */
		"virDomainGetConnect",
		"virDomainSnapshotGetConnect",
		"virDomainSnapshotGetDomain",
		"virInterfaceGetConnect",
		"virNetworkGetConnect",
		"virSecretGetConnect",
		"virStoragePoolGetConnect",
		"virStorageVolGetConnect",

		/* Only needed at C level */
		"virTypedParamsAddBoolean",
		"virTypedParamsAddDouble",
		"virTypedParamsAddFromString",
		"virTypedParamsAddInt",
		"virTypedParamsAddLLong",
		"virTypedParamsAddString",
		"virTypedParamsAddStringList",
		"virTypedParamsAddUInt",
		"virTypedParamsAddULLong",
		"virTypedParamsGet",
		"virTypedParamsGetBoolean",
		"virTypedParamsGetDouble",
		"virTypedParamsGetInt",
		"virTypedParamsGetLLong",
		"virTypedParamsGetString",
		"virTypedParamsGetUInt",
		"virTypedParamsGetULLong",
		"virTypedParamsFree",
	}

	ignoreMacros = []string{
		/* Can't be used as they contain a C format string
		 * that is not supported in go */
		"VIR_DOMAIN_TUNABLE_CPU_IOTHREADSPIN",
		"VIR_DOMAIN_TUNABLE_CPU_VCPUPIN",

		/* Compat defines for obsolete types */
		"_virBlkioParameter",
		"_virMemoryParameter",
		"_virSchedParameter",

		/* Remaped to a funcs at Go level */
		"LIBVIR_CHECK_VERSION",
		"VIR_NODEINFO_MAXCPUS",

		/* Obsoleted by VIR_TYPED_PARAM_FIELD_LENGTH */
		"VIR_DOMAIN_BLKIO_FIELD_LENGTH",
		"VIR_DOMAIN_BLOCK_STATS_FIELD_LENGTH",
		"VIR_DOMAIN_MEMORY_FIELD_LENGTH",
		"VIR_DOMAIN_SCHED_FIELD_LENGTH",
		"VIR_NODE_CPU_STATS_FIELD_LENGTH",
		"VIR_NODE_MEMORY_STATS_FIELD_LENGTH",

		/* Only needed at C level */
		"VIR_COPY_CPUMAP",
		"VIR_CPU_MAPLEN",
		"VIR_CPU_USABLE",
		"VIR_CPU_USED",
		"VIR_DOMAIN_EVENT_CALLBACK",
		"VIR_GET_CPUMAP",
		"VIR_NETWORK_EVENT_CALLBACK",
		"VIR_NODE_DEVICE_EVENT_CALLBACK",
		"VIR_SECRET_EVENT_CALLBACK",
		"VIR_SECURITY_DOI_BUFLEN",
		"VIR_SECURITY_LABEL_BUFLEN",
		"VIR_SECURITY_MODEL_BUFLEN",
		"VIR_STORAGE_POOL_EVENT_CALLBACK",
		"VIR_UNUSE_CPU",
		"VIR_USE_CPU",
	}

	ignoreEnums = []string{

		// Deprecated in favour of VIR_TYPED_PARAM_*
		"VIR_DOMAIN_BLKIO_PARAM_BOOLEAN",
		"VIR_DOMAIN_BLKIO_PARAM_DOUBLE",
		"VIR_DOMAIN_BLKIO_PARAM_INT",
		"VIR_DOMAIN_BLKIO_PARAM_LLONG",
		"VIR_DOMAIN_BLKIO_PARAM_UINT",
		"VIR_DOMAIN_BLKIO_PARAM_ULLONG",
		"VIR_DOMAIN_MEMORY_PARAM_BOOLEAN",
		"VIR_DOMAIN_MEMORY_PARAM_DOUBLE",
		"VIR_DOMAIN_MEMORY_PARAM_INT",
		"VIR_DOMAIN_MEMORY_PARAM_LLONG",
		"VIR_DOMAIN_MEMORY_PARAM_UINT",
		"VIR_DOMAIN_MEMORY_PARAM_ULLONG",
		"VIR_DOMAIN_SCHED_FIELD_BOOLEAN",
		"VIR_DOMAIN_SCHED_FIELD_DOUBLE",
		"VIR_DOMAIN_SCHED_FIELD_INT",
		"VIR_DOMAIN_SCHED_FIELD_LLONG",
		"VIR_DOMAIN_SCHED_FIELD_UINT",
		"VIR_DOMAIN_SCHED_FIELD_ULLONG",

		"VIR_TYPED_PARAM_STRING_OKAY",
	}
)

type CharsetISO88591er struct {
	r   io.ByteReader
	buf *bytes.Buffer
}

func NewCharsetISO88591(r io.Reader) *CharsetISO88591er {
	buf := bytes.Buffer{}
	return &CharsetISO88591er{r.(io.ByteReader), &buf}
}

func (cs *CharsetISO88591er) Read(p []byte) (n int, err error) {
	for _ = range p {
		if r, err := cs.r.ReadByte(); err != nil {
			break
		} else {
			cs.buf.WriteRune(rune(r))
		}
	}
	return cs.buf.Read(p)
}

func isCharset(charset string, names []string) bool {
	charset = strings.ToLower(charset)
	for _, n := range names {
		if charset == strings.ToLower(n) {
			return true
		}
	}
	return false
}

func IsCharsetISO88591(charset string) bool {
	// http://www.iana.org/assignments/character-sets
	// (last updated 2010-11-04)
	names := []string{
		// Name
		"ISO_8859-1:1987",
		// Alias (preferred MIME name)
		"ISO-8859-1",
		// Aliases
		"iso-ir-100",
		"ISO_8859-1",
		"latin1",
		"l1",
		"IBM819",
		"CP819",
		"csISOLatin1",
	}
	return isCharset(charset, names)
}

func CharsetReader(charset string, input io.Reader) (io.Reader, error) {
	if IsCharsetISO88591(charset) {
		return NewCharsetISO88591(input), nil
	}
	return input, nil
}

type APIExport struct {
	Type   string `xml:"type,attr"`
	Symbol string `xml:"symbol,attr"`
}

type APIFile struct {
	Name    string      `xml:"name,attr"`
	Exports []APIExport `xml:"exports"`
}

type API struct {
	XMLName xml.Name  `xml:"api"`
	Files   []APIFile `xml:"files>file"`
}

func GetAPIPath(varname, modname string) string {
	cmd := exec.Command("pkg-config", "--variable="+varname, modname)

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	return strings.TrimSpace(string(cmdOutput.Bytes()))
}

func GetAPI(path string) *API {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	decoder := xml.NewDecoder(file)
	decoder.CharsetReader = CharsetReader
	var api API
	err = decoder.Decode(&api)
	if err != nil {
		panic(err)
	}

	return &api
}

func GetSourceFiles() []string {
	files, _ := ioutil.ReadDir(".")

	src := make([]string, 0)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".go") &&
			!strings.HasSuffix(f.Name(), "_test.go") {
			src = append(src, f.Name())
		}
	}

	return src
}

func GetAPISymbols(api *API, funcs map[string]bool, macros map[string]bool, enums map[string]bool) {

	for _, file := range api.Files {
		for _, export := range file.Exports {
			if export.Type == "function" {
				funcs[export.Symbol] = true
			} else if export.Type == "enum" {
				if !strings.HasSuffix(export.Symbol, "_LAST") {
					enums[export.Symbol] = true
				}
			} else if export.Type == "macro" {
				macros[export.Symbol] = true
			}
		}
	}

	return
}

func ProcessFile(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	re, err := regexp.Compile("C\\.((vir|VIR|LIBVIR)[a-zA-Z0-9_]+?)(Compat|_cgo)?\\b")
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)
	symbols := make([]string, 0)
	for scanner.Scan() {
		line := scanner.Text()

		match := re.FindStringSubmatch(line)
		if match != nil {
			symbols = append(symbols, match[1])
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return symbols
}

func RecordUsage(symbols []string, funcs map[string]bool, macros map[string]bool, enums map[string]bool) {

	for _, name := range symbols {
		_, ok := funcs[name]
		if ok {
			funcs[name] = false
			continue
		}

		_, ok = macros[name]
		if ok {
			macros[name] = false
			continue
		}

		_, ok = enums[name]
		if ok {
			enums[name] = false
			continue
		}
	}
}

func ReportMissing(missingNames map[string]bool, symtype string) bool {
	missing := make([]string, 0)
	for key, value := range missingNames {
		if value {
			missing = append(missing, key)
		}
	}

	sort.Strings(missing)

	for _, name := range missing {
		fmt.Println("Missing " + symtype + " '" + name + "'")
	}

	return len(missing) != 0
}

func SetIgnores(ignores []string, symbols map[string]bool) {
	for _, name := range ignores {
		symbols[name] = false
	}
}

func TestAPICoverage(t *testing.T) {
	funcs := make(map[string]bool)
	macros := make(map[string]bool)
	enums := make(map[string]bool)

	path := GetAPIPath("libvirt_api", "libvirt")
	lxcpath := GetAPIPath("libvirt_lxc_api", "libvirt-lxc")
	qemupath := GetAPIPath("libvirt_qemu_api", "libvirt-qemu")

	api := GetAPI(path)
	lxcapi := GetAPI(lxcpath)
	qemuapi := GetAPI(qemupath)

	GetAPISymbols(api, funcs, macros, enums)
	GetAPISymbols(lxcapi, funcs, macros, enums)
	GetAPISymbols(qemuapi, funcs, macros, enums)

	SetIgnores(ignoreFuncs, funcs)
	SetIgnores(ignoreMacros, macros)
	SetIgnores(ignoreEnums, enums)

	src := GetSourceFiles()

	for _, path := range src {
		symbols := ProcessFile(path)

		RecordUsage(symbols, funcs, macros, enums)
	}

	missing := false
	if ReportMissing(funcs, "function") {
		missing = true
	}
	if ReportMissing(macros, "macro") {
		missing = true
	}
	if ReportMissing(enums, "enum") {
		missing = true
	}
	if missing {
		panic("Missing symbols found")
	}
}
