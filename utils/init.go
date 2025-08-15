package utils

import (
	"fmt"
	"github.com/everFinance/goether"
	"github.com/hymatrix/hymx/sdk"
	"github.com/permadao/goar"

	"github.com/permadao/goar/schema"
)

const (
	tokenModule    = "1i03Vpe8DljkUMBEEEvR0VmbJjvgZtP_ytZdThkVSMw"
	registryModule = "MVTil0kn5SRiJELW7W2jLZ6cBr3QUGj1nJ67I2Wi4Ps"
)

var (
	url        = "http://127.0.0.1:8080"
	prvKey     = "0x64dd2342616f385f3e8157cf7246cf394217e13e8f91b7d208e9f8b60e25ed1b" // local test keyFile
	signer, _  = goether.NewSigner(prvKey)
	bundler, _ = goar.NewBundler(signer)
	s          = sdk.NewFromBundler(url, bundler)
)

func initToken() string {
	res, err := s.SpawnAndWait(
		tokenModule,
		s.GetAddress(),
		[]schema.Tag{})
	fmt.Println(res, err)
	return res.Id
}

func initRegistry(tokenPid string) {
	res, err := s.SpawnAndWait(
		registryModule,
		s.GetAddress(),
		[]schema.Tag{
			{Name: "Token-Pid", Value: tokenPid},
		})
	fmt.Println(res, err)
}
