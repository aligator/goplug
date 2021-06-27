package actions

import (
	api0 "github.com/aligator/goplug/example/host/api"
	apackage0 "github.com/aligator/goplug/example/host/api/a_package"
	"github.com/aligator/goplug/goplug"
)

// HostActions contains the host-implementations of actions.
type HostActions struct {
	Api0AppRef *api0.App
}

type ClientActions struct {
	client *goplug.Client
}

func NewClientActions(plugin *goplug.Client) ClientActions {
	return ClientActions{
		client: plugin,
	}
}

// Make some plugin-methods available to the plugins.

func (c *ClientActions) Print(text string) error {
	return c.client.Print(text)
}

// Action implementations for host and client.

type GetRandomIntRequest struct {
	N int `json:"n"`
}

type GetRandomIntResponse struct {
	Res int `json:"res"`
}

// GetRandomInt returns, a non-negative pseudo-random number in [0,n) from the
// default Source. Returns an error if n <= 0.
func (h *HostActions) GetRandomInt(args GetRandomIntRequest, reply *GetRandomIntResponse) error {
	// Host implementation.
	res, err := h.Api0AppRef.GetRandomInt(
		args.N,
	)

	if err != nil {
		return err
	}

	*reply = GetRandomIntResponse{
		Res: res,
	}

	return nil
}

// GetRandomInt returns, a non-negative pseudo-random number in [0,n) from the
// default Source. Returns an error if n <= 0.
func (c *ClientActions) GetRandomInt(
	n int,
) (res int, err error) {
	// Calling from the plugin.
	response := GetRandomIntResponse{}
	err = c.client.Call("GetRandomInt", GetRandomIntRequest{
		N: n,
	}, &response)
	return response.Res, err
}

type PrintHelloRequest struct {
}

type PrintHelloResponse struct {
}

// PrintHello to stdout.
func (h *HostActions) PrintHello(args PrintHelloRequest, reply *PrintHelloResponse) error {
	// Host implementation.
	err := h.Api0AppRef.PrintHello()

	if err != nil {
		return err
	}

	return nil
}

// PrintHello to stdout.
func (c *ClientActions) PrintHello() error {
	// Calling from the plugin.
	response := PrintHelloResponse{}
	err := c.client.Call("PrintHello", PrintHelloRequest{}, &response)
	return err
}

type WithPointerRequest struct {
	Val *int `json:"val"`
}

type WithPointerResponse struct {
	Res0 *int `json:"res0"`
}

func (h *HostActions) WithPointer(args WithPointerRequest, reply *WithPointerResponse) error {
	// Host implementation.
	res0, err := h.Api0AppRef.WithPointer(
		args.Val,
	)

	if err != nil {
		return err
	}

	*reply = WithPointerResponse{
		Res0: res0,
	}

	return nil
}

func (c *ClientActions) WithPointer(
	val *int,
) (res0 *int, err error) {
	// Calling from the plugin.
	response := WithPointerResponse{}
	err = c.client.Call("WithPointer", WithPointerRequest{
		Val: val,
	}, &response)
	return response.Res0, err
}

type WithStructRequest struct {
	Val api0.AStruct `json:"val"`
}

type WithStructResponse struct {
	Res0 api0.AStruct `json:"res0"`
}

func (h *HostActions) WithStruct(args WithStructRequest, reply *WithStructResponse) error {
	// Host implementation.
	res0, err := h.Api0AppRef.WithStruct(
		args.Val,
	)

	if err != nil {
		return err
	}

	*reply = WithStructResponse{
		Res0: res0,
	}

	return nil
}

func (c *ClientActions) WithStruct(
	val api0.AStruct,
) (res0 api0.AStruct, err error) {
	// Calling from the plugin.
	response := WithStructResponse{}
	err = c.client.Call("WithStruct", WithStructRequest{
		Val: val,
	}, &response)
	return response.Res0, err
}

type WithPointerToStructRequest struct {
	Val *api0.AStruct `json:"val"`
}

type WithPointerToStructResponse struct {
	Res0 *api0.AStruct `json:"res0"`
}

func (h *HostActions) WithPointerToStruct(args WithPointerToStructRequest, reply *WithPointerToStructResponse) error {
	// Host implementation.
	res0, err := h.Api0AppRef.WithPointerToStruct(
		args.Val,
	)

	if err != nil {
		return err
	}

	*reply = WithPointerToStructResponse{
		Res0: res0,
	}

	return nil
}

func (c *ClientActions) WithPointerToStruct(
	val *api0.AStruct,
) (res0 *api0.AStruct, err error) {
	// Calling from the plugin.
	response := WithPointerToStructResponse{}
	err = c.client.Call("WithPointerToStruct", WithPointerToStructRequest{
		Val: val,
	}, &response)
	return response.Res0, err
}

type WithStructFromPackageRequest struct {
	Val apackage0.AStruct `json:"val"`
}

type WithStructFromPackageResponse struct {
	Res0 apackage0.AStruct `json:"res0"`
}

func (h *HostActions) WithStructFromPackage(args WithStructFromPackageRequest, reply *WithStructFromPackageResponse) error {
	// Host implementation.
	res0, err := h.Api0AppRef.WithStructFromPackage(
		args.Val,
	)

	if err != nil {
		return err
	}

	*reply = WithStructFromPackageResponse{
		Res0: res0,
	}

	return nil
}

func (c *ClientActions) WithStructFromPackage(
	val apackage0.AStruct,
) (res0 apackage0.AStruct, err error) {
	// Calling from the plugin.
	response := WithStructFromPackageResponse{}
	err = c.client.Call("WithStructFromPackage", WithStructFromPackageRequest{
		Val: val,
	}, &response)
	return response.Res0, err
}

type WithPointerToStructFromPackageRequest struct {
	Val *apackage0.AStruct `json:"val"`
}

type WithPointerToStructFromPackageResponse struct {
	Res0 *apackage0.AStruct `json:"res0"`
}

func (h *HostActions) WithPointerToStructFromPackage(args WithPointerToStructFromPackageRequest, reply *WithPointerToStructFromPackageResponse) error {
	// Host implementation.
	res0, err := h.Api0AppRef.WithPointerToStructFromPackage(
		args.Val,
	)

	if err != nil {
		return err
	}

	*reply = WithPointerToStructFromPackageResponse{
		Res0: res0,
	}

	return nil
}

func (c *ClientActions) WithPointerToStructFromPackage(
	val *apackage0.AStruct,
) (res0 *apackage0.AStruct, err error) {
	// Calling from the plugin.
	response := WithPointerToStructFromPackageResponse{}
	err = c.client.Call("WithPointerToStructFromPackage", WithPointerToStructFromPackageRequest{
		Val: val,
	}, &response)
	return response.Res0, err
}
