package provider

import "time"

#Provider: close({
	version:     "provider/v1"
	name:        string & =~"^[a-z][a-z0-9._-]*$"
	description: string & !=""
	command: [string & !="", ...string & !=""]
	actions: [=~"^[a-z][a-z0-9._-]*$"]: #Action
	requires?: #Requirements
	defaults?: #Defaults
})

#Action: close({
	description: string & !=""
	argv?: [...string]
	env?: [=~"^[A-Za-z_][A-Za-z0-9_]*$"]: string
})

#Requirements: close({
	commands?: [...string & !=""]
	environment?: [...string & =~"^[A-Za-z_][A-Za-z0-9_]*$"]
	paths?: [...string & !=""]
})

#Defaults: close({
	timeout?:  time.Duration
	priority?: int
})
