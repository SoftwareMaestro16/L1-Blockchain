package app

import "github.com/sovereign-l1/l1/app/abcihandlers"

type VoteExtensionHandler = abcihandlers.VoteExtensionHandler
type VoteExtension = abcihandlers.VoteExtension

func NewVoteExtensionHandler() *VoteExtensionHandler {
	return abcihandlers.NewVoteExtensionHandler()
}
