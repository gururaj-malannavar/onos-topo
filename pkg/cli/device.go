// Copyright 2019-present Open Networking Foundation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"bytes"
	"context"
	"fmt"
	"github.com/onosproject/onos-topo/pkg/northbound/device"
	"github.com/spf13/cobra"
	"io"
	"os"
	"text/tabwriter"
	"time"
)

func getGetDeviceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "device <id>",
		Aliases: []string{"devices"},
		Args:    cobra.MaximumNArgs(1),
		Short:   "Get a device",
		Run:     runGetDeviceCommand,
	}
	cmd.Flags().BoolP("verbose", "v", false, "whether to print the device with verbose output")
	cmd.Flags().Bool("no-headers", false, "disables output headers")
	return cmd
}

func runGetDeviceCommand(cmd *cobra.Command, args []string) {
	verbose, _ := cmd.Flags().GetBool("verbose")
	noHeaders, _ := cmd.Flags().GetBool("no-headers")

	conn := getConnection()
	defer conn.Close()

	client := device.NewDeviceServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if len(args) == 0 {
		stream, err := client.List(ctx, &device.ListRequest{})
		if err != nil {
			ExitWithError(ExitBadConnection, err)
		}

		writer := new(tabwriter.Writer)
		writer.Init(os.Stdout, 0, 0, 3, ' ', tabwriter.FilterHTML)

		if !noHeaders {
			if verbose {
				fmt.Fprintln(writer, "ID\tADDRESS\tVERSION\tUSER\tPASSWORD\tATTRIBUTES")
			} else {
				fmt.Fprintln(writer, "ID\tADDRESS\tVERSION")
			}
		}

		for {
			response, err := stream.Recv()
			if err == io.EOF {
				break
			} else if err != nil {
				ExitWithError(ExitError, err)
			}

			dev := response.Device
			if verbose {
				attributesBuf := bytes.Buffer{}
				for key, attribute := range dev.Attributes {
					attributesBuf.WriteString(key)
					attributesBuf.WriteString(": ")
					attributesBuf.WriteString(attribute)
					attributesBuf.WriteString(", ")
				}
				fmt.Fprintln(writer, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s", dev.ID, dev.Address, dev.Version,
					dev.Credentials.User, dev.Credentials.Password, attributesBuf.String()))
			} else {
				fmt.Fprintln(writer, fmt.Sprintf("%s\t%s\t%s", dev.ID, dev.Address, dev.Version))
			}
		}
		writer.Flush()
	} else {
		response, err := client.Get(ctx, &device.GetRequest{
			ID: device.ID(args[0]),
		})
		if err != nil {
			ExitWithError(ExitBadConnection, err)
		}

		dev := response.Device

		writer := new(tabwriter.Writer)
		writer.Init(os.Stdout, 0, 0, 3, ' ', tabwriter.FilterHTML)
		fmt.Fprintln(writer, fmt.Sprintf("ID\t%s", dev.ID))
		fmt.Fprintln(writer, fmt.Sprintf("ADDRESS\t%s", dev.Address))
		fmt.Fprintln(writer, fmt.Sprintf("VERSION\t%s", dev.Version))

		if verbose {
			fmt.Fprintln(writer, fmt.Sprintf("USER\t%s", dev.Credentials.User))
			fmt.Fprintln(writer, fmt.Sprintf("PASSWORD\t%s", dev.Credentials.Password))
		}
		writer.Flush()
	}
}

func getAddDeviceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "device <id> [args]",
		Aliases: []string{"devices"},
		Args:    cobra.ExactArgs(1),
		Short:   "Add a device",
		Run:     runAddDeviceCommand,
	}
	cmd.Flags().StringP("type", "t", "", "the type of the device")
	cmd.Flags().StringP("role", "r", "", "the device role")
	cmd.Flags().StringP("target", "g", "", "the device target name")
	cmd.Flags().StringP("address", "a", "", "the address of the device")
	cmd.Flags().StringP("user", "u", "", "the device username")
	cmd.Flags().StringP("password", "p", "", "the device password")
	cmd.Flags().StringP("version", "v", "", "the device software version")
	cmd.Flags().String("key", "", "the TLS key")
	cmd.Flags().String("cert", "", "the TLS certificate")
	cmd.Flags().String("ca-cert", "", "the TLS CA certificate")
	cmd.Flags().Bool("plain", false, "whether to connect over a plaintext connection")
	cmd.Flags().Bool("insecure", false, "whether to enable skip verification")
	cmd.Flags().Duration("timeout", 5*time.Second, "the device connection timeout")
	cmd.Flags().StringToString("attributes", map[string]string{}, "an arbitrary mapping of device attributes")

	_ = cmd.MarkFlagRequired("version")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}

func runAddDeviceCommand(cmd *cobra.Command, args []string) {
	id := args[0]
	deviceType, _ := cmd.Flags().GetString("type")
	deviceRole, _ := cmd.Flags().GetString("role")
	deviceTarget, _ := cmd.Flags().GetString("target")
	address, _ := cmd.Flags().GetString("address")
	user, _ := cmd.Flags().GetString("user")
	password, _ := cmd.Flags().GetString("password")
	version, _ := cmd.Flags().GetString("version")
	key, _ := cmd.Flags().GetString("key")
	cert, _ := cmd.Flags().GetString("cert")
	caCert, _ := cmd.Flags().GetString("ca-cert")
	plain, _ := cmd.Flags().GetBool("plain")
	insecure, _ := cmd.Flags().GetBool("insecure")
	timeout, _ := cmd.Flags().GetDuration("timeout")
	attributes, _ := cmd.Flags().GetStringToString("attributes")

	// Target defaults to the ID
	if deviceTarget == "" {
		deviceTarget = id
	}

	conn := getConnection()
	defer conn.Close()

	client := device.NewDeviceServiceClient(conn)

	dev := &device.Device{
		ID:      device.ID(id),
		Type:    device.Type(deviceType),
		Role:    device.Role(deviceRole),
		Address: address,
		Target:  deviceTarget,
		Version: version,
		Timeout: &timeout,
		Credentials: device.Credentials{
			User:     user,
			Password: password,
		},
		TLS: device.TlsConfig{
			Cert:     cert,
			Key:      key,
			CaCert:   caCert,
			Plain:    plain,
			Insecure: insecure,
		},
		Attributes: attributes,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := client.Add(ctx, &device.AddRequest{
		Device: dev,
	})
	if err != nil {
		ExitWithError(ExitBadConnection, err)
	} else {
		ExitWithOutput("Added device %s", id)
	}
}

func getUpdateDeviceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "device <id> [args]",
		Aliases: []string{"devices"},
		Args:    cobra.ExactArgs(1),
		Short:   "Update a device",
		Run:     runUpdateDeviceCommand,
	}
	cmd.Flags().StringP("type", "t", "", "the type of the device")
	cmd.Flags().StringP("role", "r", "", "the device role")
	cmd.Flags().StringP("target", "g", "", "the device target name")
	cmd.Flags().StringP("address", "a", "", "the address of the device")
	cmd.Flags().StringP("user", "u", "", "the device username")
	cmd.Flags().StringP("password", "p", "", "the device password")
	cmd.Flags().StringP("version", "v", "", "the device software version")
	cmd.Flags().String("key", "", "the TLS key")
	cmd.Flags().String("cert", "", "the TLS certificate")
	cmd.Flags().String("ca-cert", "", "the TLS CA certificate")
	cmd.Flags().Bool("plain", false, "whether to connect over a plaintext connection")
	cmd.Flags().Bool("insecure", false, "whether to enable skip verification")
	cmd.Flags().Duration("timeout", 30*time.Second, "the device connection timeout")
	cmd.Flags().StringToString("attributes", map[string]string{}, "an arbitrary mapping of device attributes")
	return cmd
}

func runUpdateDeviceCommand(cmd *cobra.Command, args []string) {
	id := args[0]

	conn := getConnection()
	defer conn.Close()

	client := device.NewDeviceServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	response, err := client.Get(ctx, &device.GetRequest{
		ID: device.ID(id),
	})
	if err != nil {
		ExitWithError(ExitBadConnection, err)
	}

	cancel()
	dvc := response.Device

	if cmd.Flags().Changed("type") {
		deviceType, _ := cmd.Flags().GetString("type")
		dvc.Type = device.Type(deviceType)
	}
	if cmd.Flags().Changed("target") {
		deviceTarget, _ := cmd.Flags().GetString("target")
		dvc.Target = deviceTarget
	}
	if cmd.Flags().Changed("role") {
		deviceRole, _ := cmd.Flags().GetString("role")
		dvc.Role = device.Role(deviceRole)
	}
	if cmd.Flags().Changed("address") {
		address, _ := cmd.Flags().GetString("address")
		dvc.Address = address
	}
	if cmd.Flags().Changed("user") {
		user, _ := cmd.Flags().GetString("user")
		dvc.Credentials.User = user
	}
	if cmd.Flags().Changed("password") {
		password, _ := cmd.Flags().GetString("password")
		dvc.Credentials.Password = password
	}
	if cmd.Flags().Changed("version") {
		version, _ := cmd.Flags().GetString("version")
		dvc.Version = version
	}
	if cmd.Flags().Changed("key") {
		key, _ := cmd.Flags().GetString("key")
		dvc.TLS.Key = key
	}
	if cmd.Flags().Changed("cert") {
		cert, _ := cmd.Flags().GetString("cert")
		dvc.TLS.Cert = cert
	}
	if cmd.Flags().Changed("ca-cert") {
		caCert, _ := cmd.Flags().GetString("ca-cert")
		dvc.TLS.CaCert = caCert
	}
	if cmd.Flags().Changed("plain") {
		plain, _ := cmd.Flags().GetBool("plain")
		dvc.TLS.Plain = plain
	}
	if cmd.Flags().Changed("insecure") {
		insecure, _ := cmd.Flags().GetBool("insecure")
		dvc.TLS.Insecure = insecure
	}
	if cmd.Flags().Changed("timeout") {
		timeout, _ := cmd.Flags().GetDuration("timeout")
		dvc.Timeout = &timeout
	}
	if cmd.Flags().Changed("attributes") {
		attributes, _ := cmd.Flags().GetStringToString("attributes")
		dvc.Attributes = attributes
	}

	ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err = client.Update(ctx, &device.UpdateRequest{
		Device: dvc,
	})
	if err != nil {
		ExitWithError(ExitBadConnection, err)
	} else {
		ExitWithOutput("Updated device %s", id)
	}
}

func getRemoveDeviceCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "device <id> [args]",
		Aliases: []string{"devices"},
		Args:    cobra.ExactArgs(1),
		Short:   "Remove a device",
		Run:     runRemoveDeviceCommand,
	}
}

func runRemoveDeviceCommand(cmd *cobra.Command, args []string) {
	id := args[0]

	conn := getConnection()
	defer conn.Close()

	client := device.NewDeviceServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := client.Remove(ctx, &device.RemoveRequest{
		Device: &device.Device{
			ID: device.ID(id),
		},
	})
	if err != nil {
		ExitWithError(ExitBadConnection, err)
	} else {
		ExitWithOutput("Removed device %s", id)
	}
}

func getWatchDeviceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "device <id> [args]",
		Aliases: []string{"devices"},
		Args:    cobra.MaximumNArgs(1),
		Short:   "Watch for device changes",
		Run:     runWatchDeviceCommand,
	}
	cmd.Flags().BoolP("verbose", "v", false, "whether to print the device with verbose output")
	cmd.Flags().Bool("no-headers", false, "disables output headers")
	return cmd
}

func runWatchDeviceCommand(cmd *cobra.Command, args []string) {
	var id string
	if len(args) > 0 {
		id = args[0]
	}

	verbose, _ := cmd.Flags().GetBool("verbose")
	noHeaders, _ := cmd.Flags().GetBool("no-headers")

	conn := getConnection()
	defer conn.Close()

	client := device.NewDeviceServiceClient(conn)

	stream, err := client.List(context.Background(), &device.ListRequest{
		Subscribe: true,
	})
	if err != nil {
		ExitWithError(ExitBadConnection, err)
	}

	writer := new(tabwriter.Writer)
	writer.Init(os.Stdout, 0, 0, 3, ' ', tabwriter.FilterHTML)

	if !noHeaders {
		if verbose {
			fmt.Fprintln(writer, "EVENT\tID\tADDRESS\tVERSION\tUSER\tPASSWORD")
		} else {
			fmt.Fprintln(writer, "EVENT\tID\tADDRESS\tVERSION")
		}
		writer.Flush()
	}

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			ExitWithSuccess()
		} else if err != nil {
			ExitWithError(ExitError, err)
		}

		dev := response.Device
		if id != "" && dev.ID != device.ID(id) {
			continue
		}

		if verbose {
			fmt.Fprintln(writer, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s", response.Type, dev.ID, dev.Address, dev.Version, dev.Credentials.User, dev.Credentials.Password))
		} else {
			fmt.Fprintln(writer, fmt.Sprintf("%s\t%s\t%s\t%s", response.Type, dev.ID, dev.Address, dev.Version))
		}
		writer.Flush()
	}
}
