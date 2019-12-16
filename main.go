package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/dmacvicar/terraform-provider-libvirt/libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	libvirtgo "github.com/libvirt/libvirt-go"
)

var version = "was not built correctly" // set via the Makefile

func main() {
	versionFlag := flag.Bool("version", false, "print version information and exit")
	flag.Parse()
	if *versionFlag {
		err := printVersion(os.Stdout)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	defer libvirt.CleanupLibvirtConnections()

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: libvirt.Provider,
	})
}

func printVersion(writer io.Writer) error {
	fmt.Fprintf(writer, "%s %s\n", os.Args[0], version)

	fmt.Fprintf(writer, "Compiled against library: libvirt %s\n", parseVersion(libvirtgo.VERSION_NUMBER))

	libvirtVersion, err := libvirtgo.GetVersion()
	if err != nil {
		return err
	}
	fmt.Fprintf(writer, "Using library: libvirt %s\n", parseVersion(libvirtVersion))

	conn, err := libvirtgo.NewConnect("")
	if err != nil {
		return err
	}
	defer conn.Close()

	hvType, err := conn.GetType()
	if err != nil {
		return err
	}
	libvirtVersion, err = conn.GetVersion()
	if err != nil {
		return err
	}
	fmt.Fprintf(writer, "Running hypervisor: %s %s\n", hvType, parseVersion(libvirtVersion))

	libvirtVersion, err = conn.GetLibVersion()
	if err != nil {
		return err
	}
	fmt.Fprintf(writer, "Running against daemon: %s\n", parseVersion(libvirtVersion))

	return nil
}

func parseVersion(version uint32) string {
	release := version % 1000
	version /= 1000
	minor := version % 1000
	major := version / 1000
	return fmt.Sprintf("%d.%d.%d", major, minor, release)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
